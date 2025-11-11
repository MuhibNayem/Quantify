package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/storage"
)

type ReportingService struct {
	repo                 *repository.ReportsRepository
	uploader             storage.Uploader
	jobRepo              *repository.JobRepository
	salesTrendsTTL       time.Duration
	inventoryTurnoverTTL time.Duration
	profitMarginTTL      time.Duration
}

func NewReportingService(repo *repository.ReportsRepository, uploader storage.Uploader, jobRepo *repository.JobRepository, cfg *config.Config) *ReportingService {
	return &ReportingService{
		repo:                 repo,
		uploader:             uploader,
		jobRepo:              jobRepo,
		salesTrendsTTL:       cfg.SalesTrendsCacheTTL,
		inventoryTurnoverTTL: cfg.InventoryTurnoverCacheTTL,
		profitMarginTTL:      cfg.ProfitMarginCacheTTL,
	}
}

func (s *ReportingService) GenerateReport(job *domain.Job) error {
	var buffer bytes.Buffer
	var fileType string

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(job.Payload), &params); err != nil {
		return err
	}

	switch job.Type {
	case "sales_trends_csv":
		fileType = "csv"
		startDate, _ := time.Parse(time.RFC3339, params["startDate"].(string))
		endDate, _ := time.Parse(time.RFC3339, params["endDate"].(string))
		var categoryID, locationID *uint
		if catID, ok := params["categoryId"]; ok && catID != nil {
			val := uint(catID.(float64))
			categoryID = &val
		}
		if locID, ok := params["locationId"]; ok && locID != nil {
			val := uint(locID.(float64))
			locationID = &val
		}
		groupBy := params["groupBy"].(string)
		err := s.ExportSalesTrendsReport(&buffer, startDate, endDate, categoryID, locationID, groupBy)
		if err != nil {
			return err
		}
	case "sales_trends_pdf":
		fileType = "pdf"
		startDate, _ := time.Parse(time.RFC3339, params["startDate"].(string))
		endDate, _ := time.Parse(time.RFC3339, params["endDate"].(string))
		var categoryID, locationID *uint
		if catID, ok := params["categoryId"]; ok && catID != nil {
			val := uint(catID.(float64))
			categoryID = &val
		}
		if locID, ok := params["locationId"]; ok && locID != nil {
			val := uint(locID.(float64))
			locationID = &val
		}
		groupBy := params["groupBy"].(string)
		err := s.ExportSalesTrendsReportPDF(&buffer, startDate, endDate, categoryID, locationID, groupBy)
		if err != nil {
			return err
		}
	case "sales_trends_excel":
		fileType = "xlsx"
		startDate, _ := time.Parse(time.RFC3339, params["startDate"].(string))
		endDate, _ := time.Parse(time.RFC3339, params["endDate"].(string))
		var categoryID, locationID *uint
		if catID, ok := params["categoryId"]; ok && catID != nil {
			val := uint(catID.(float64))
			categoryID = &val
		}
		if locID, ok := params["locationId"]; ok && locID != nil {
			val := uint(locID.(float64))
			locationID = &val
		}
		groupBy := params["groupBy"].(string)
		err := s.ExportSalesTrendsReportExcel(&buffer, startDate, endDate, categoryID, locationID, groupBy)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown report type: %s", job.Type)
	}

	fileName := fmt.Sprintf("report-%d.%s", job.ID, fileType)
	uploadInfo, err := s.uploader.UploadFile("reports", fileName, &buffer, int64(buffer.Len()))
	if err != nil {
		return err
	}

	job.Status = "COMPLETED"
	job.Result = uploadInfo.Location
	now := time.Now()
	job.LastAttemptAt = &now
	return s.jobRepo.UpdateJob(job)
}

func (s *ReportingService) GetSalesTrendsReport(startDate, endDate time.Time, categoryID, locationID *uint, groupBy string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("sales_trends:%s:%s:%v:%v:%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), categoryID, locationID, groupBy)
	cachedReport, err := repository.GetCache(cacheKey)
	if err == nil && cachedReport != "" {
		var reportData map[string]interface{}
		err = json.Unmarshal([]byte(cachedReport), &reportData)
		if err == nil {
			return reportData, nil
		}
	}

	salesTrends, topSellingProducts, err := s.repo.GetSalesTrends(startDate, endDate, categoryID, locationID, groupBy)
	if err != nil {
		return nil, err
	}

	var totalSales float64
	for _, trend := range salesTrends {
		totalSales += trend.TotalSales
	}

	days := endDate.Sub(startDate).Hours() / 24
	if days == 0 {
		days = 1
	}
	averageDailySales := totalSales / days

	reportData := map[string]interface{}{
		"period":             fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"totalSales":         totalSales,
		"averageDailySales":  averageDailySales,
		"salesTrends":        salesTrends,
		"topSellingProducts": topSellingProducts,
	}

	if s.salesTrendsTTL > 0 {
		reportJSON, err := json.Marshal(reportData)
		if err != nil {
			logrus.WithFields(logrus.Fields{"cacheKey": cacheKey}).WithError(err).
				Error("Failed to marshal sales trends report for caching")
		} else if err := repository.SetCache(cacheKey, string(reportJSON), s.salesTrendsTTL); err != nil {
			logrus.WithFields(logrus.Fields{"cacheKey": cacheKey}).WithError(err).
				Error("Failed to cache sales trends report")
		}
	}

	return reportData, nil
}

func (s *ReportingService) ExportSalesTrendsReport(writer io.Writer, startDate, endDate time.Time, categoryID, locationID *uint, groupBy string) error {
	reportData, err := s.GetSalesTrendsReport(startDate, endDate, categoryID, locationID, groupBy)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	csvWriter.Write([]string{"Date", "Total Sales"})

	// Write data
	salesTrends := reportData["salesTrends"].([]repository.SalesTrend)
	for _, trend := range salesTrends {
		date := trend.Date.Format("2006-01-02")
		totalSales := strconv.FormatFloat(trend.TotalSales, 'f', 2, 64)
		csvWriter.Write([]string{date, totalSales})
	}

	return nil
}

func (s *ReportingService) ExportSalesTrendsReportPDF(writer io.Writer, startDate, endDate time.Time, categoryID, locationID *uint, groupBy string) error {
	reportData, err := s.GetSalesTrendsReport(startDate, endDate, categoryID, locationID, groupBy)
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Sales Trends Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Period: %s", reportData["period"]))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Date")
	pdf.Cell(40, 10, "Total Sales")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	salesTrends := reportData["salesTrends"].([]repository.SalesTrend)
	for _, trend := range salesTrends {
		pdf.Cell(40, 10, trend.Date.Format("2006-01-02"))
		pdf.Cell(40, 10, strconv.FormatFloat(trend.TotalSales, 'f', 2, 64))
		pdf.Ln(10)
	}

	return pdf.Output(writer)
}

func (s *ReportingService) ExportSalesTrendsReportExcel(writer io.Writer, startDate, endDate time.Time, categoryID, locationID *uint, groupBy string) error {
	reportData, err := s.GetSalesTrendsReport(startDate, endDate, categoryID, locationID, groupBy)
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	sheetName := "Sales Trends"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	f.SetCellValue(sheetName, "A1", "Date")
	f.SetCellValue(sheetName, "B1", "Total Sales")

	salesTrends := reportData["salesTrends"].([]repository.SalesTrend)
	for i, trend := range salesTrends {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), trend.Date.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), trend.TotalSales)
	}

	return f.Write(writer)
}

func (s *ReportingService) GetInventoryTurnoverReport(startDate, endDate time.Time, categoryID, locationID *uint) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("inventory_turnover:%s:%s:%v:%v", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), categoryID, locationID)
	if cached, err := repository.GetCache(cacheKey); err == nil && cached != "" {
		var reportData map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &reportData); err == nil {
			return reportData, nil
		}
	}

	costOfGoodsSold, averageInventoryValue, err := s.repo.GetInventoryTurnover(startDate, endDate, categoryID, locationID)
	if err != nil {
		return nil, err
	}

	var inventoryTurnoverRate float64
	if averageInventoryValue > 0 {
		inventoryTurnoverRate = costOfGoodsSold / averageInventoryValue
	}

	reportData := map[string]interface{}{
		"period":                fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"totalCostOfGoodsSold":  costOfGoodsSold,
		"averageInventoryValue": averageInventoryValue,
		"inventoryTurnoverRate": inventoryTurnoverRate,
	}

	if s.inventoryTurnoverTTL > 0 {
		reportJSON, err := json.Marshal(reportData)
		if err != nil {
			logrus.WithFields(logrus.Fields{"cacheKey": cacheKey}).WithError(err).
				Error("Failed to marshal inventory turnover report for caching")
		} else if err := repository.SetCache(cacheKey, string(reportJSON), s.inventoryTurnoverTTL); err != nil {
			logrus.WithFields(logrus.Fields{"cacheKey": cacheKey}).WithError(err).
				Error("Failed to cache inventory turnover report")
		}
	}

	return reportData, nil
}

func (s *ReportingService) GetProfitMarginReport(startDate, endDate time.Time, categoryID, locationID *uint) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("profit_margin:%s:%s:%v:%v", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), categoryID, locationID)
	if cached, err := repository.GetCache(cacheKey); err == nil && cached != "" {
		var reportData map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &reportData); err == nil {
			return reportData, nil
		}
	}

	totalRevenue, totalCost, err := s.repo.GetProfitMargin(startDate, endDate, categoryID, locationID)
	if err != nil {
		return nil, err
	}

	grossProfit := totalRevenue - totalCost
	var grossProfitMargin float64
	if totalRevenue > 0 {
		grossProfitMargin = grossProfit / totalRevenue
	}

	reportData := map[string]interface{}{
		"period":            fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"totalRevenue":      totalRevenue,
		"totalCost":         totalCost,
		"grossProfit":       grossProfit,
		"grossProfitMargin": grossProfitMargin,
	}

	if s.profitMarginTTL > 0 {
		reportJSON, err := json.Marshal(reportData)
		if err != nil {
			logrus.WithFields(logrus.Fields{"cacheKey": cacheKey}).WithError(err).
				Error("Failed to marshal profit margin report for caching")
		} else if err := repository.SetCache(cacheKey, string(reportJSON), s.profitMarginTTL); err != nil {
			logrus.WithFields(logrus.Fields{"cacheKey": cacheKey}).WithError(err).
				Error("Failed to cache profit margin report")
		}
	}

	return reportData, nil
}

func (s *ReportingService) GenerateDailySalesSummary() {
	yesterday := time.Now().AddDate(0, 0, -1)
	totalSales, err := s.repo.GetDailySalesSummary(yesterday)
	if err != nil {
		logrus.Errorf("Failed to generate daily sales summary: %v", err)
		return
	}

	logrus.Infof("Daily sales summary generated for %s: Total Sales = $%.2f", yesterday.Format("2006-01-02"), totalSales)
}
