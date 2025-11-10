package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/storage"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type ReportingService struct {
	repo     *repository.ReportsRepository
	uploader storage.Uploader
}

func NewReportingService(repo *repository.ReportsRepository, uploader storage.Uploader) *ReportingService {
	return &ReportingService{repo: repo, uploader: uploader}
}

type ReportJob struct {
	JobID      string      `json:"jobId"`
	ReportType string      `json:"reportType"`
	Params     interface{} `json:"params"`
	Status     string      `json:"status"`
	FileURL    string      `json:"fileUrl"`
}

func (s *ReportingService) RequestReport(reportType string, params interface{}) (string, error) {
	jobID := uuid.New().String()
	job := ReportJob{
		JobID:      jobID,
		ReportType: reportType,
		Params:     params,
		Status:     "QUEUED",
	}

	err := message_broker.Publish("inventory", "reporting", job)
	if err != nil {
		return "", err
	}

	// Store job status in Redis
	payload, err := json.Marshal(job)
	if err != nil {
		return "", err
	}
	err = repository.SetCache("report_job:"+jobID, string(payload), 3600)
	if err != nil {
		return "", err
	}

	return jobID, nil
}

func (s *ReportingService) GenerateReport(job *ReportJob) error {
	var buffer bytes.Buffer
	var fileType string

	switch job.ReportType {
	case "sales_trends_csv":
		fileType = "csv"
		params := job.Params.(map[string]interface{})
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
		params := job.Params.(map[string]interface{})
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
		params := job.Params.(map[string]interface{})
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
		return fmt.Errorf("unknown report type: %s", job.ReportType)
	}

	fileName := fmt.Sprintf("report-%s.%s", job.JobID, fileType)
	uploadInfo, err := s.uploader.UploadFile("reports", fileName, &buffer, int64(buffer.Len()))
	if err != nil {
		return err
	}

	job.Status = "COMPLETED"
	job.FileURL = uploadInfo.Location
	payload, _ := json.Marshal(job)
	return repository.SetCache("report_job:"+job.JobID, string(payload), 3600)
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

	reportJSON, _ := json.Marshal(reportData)
	repository.SetCache(cacheKey, string(reportJSON), 3600)

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
	costOfGoodsSold, averageInventoryValue, err := s.repo.GetInventoryTurnover(startDate, endDate, categoryID, locationID)
	if err != nil {
		return nil, err
	}

	var inventoryTurnoverRate float64
	if averageInventoryValue > 0 {
		inventoryTurnoverRate = costOfGoodsSold / averageInventoryValue
	}

	reportData := map[string]interface{}{
		"period":                  fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		"totalCostOfGoodsSold":    costOfGoodsSold,
		"averageInventoryValue":   averageInventoryValue,
		"inventoryTurnoverRate":   inventoryTurnoverRate,
	}

	return reportData, nil
}

func (s *ReportingService) GetProfitMarginReport(startDate, endDate time.Time, categoryID, locationID *uint) (map[string]interface{}, error) {
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
