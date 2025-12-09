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
	"inventory/backend/internal/websocket"
)

type ReportingService struct {
	repo                 *repository.ReportsRepository
	uploader             storage.Uploader
	jobRepo              *repository.JobRepository
	hub                  *websocket.Hub
	salesTrendsTTL       time.Duration
	inventoryTurnoverTTL time.Duration
	profitMarginTTL      time.Duration
}

func NewReportingService(repo *repository.ReportsRepository, uploader storage.Uploader, jobRepo *repository.JobRepository, hub *websocket.Hub, cfg *config.Config) *ReportingService {
	return &ReportingService{
		repo:                 repo,
		uploader:             uploader,
		jobRepo:              jobRepo,
		hub:                  hub,
		salesTrendsTTL:       cfg.SalesTrendsCacheTTL,
		inventoryTurnoverTTL: cfg.InventoryTurnoverCacheTTL,
		profitMarginTTL:      cfg.ProfitMarginCacheTTL,
	}
}

// NotifyReportUpdate triggers a real-time update for a specific report type.
// It broadcasts a message to all connected clients with permission to view reports.
// NotifyReportUpdate triggers a real-time update for a specific report type.
// It fetches the latest data for that report and broadcasts it to all connected clients.
func (s *ReportingService) NotifyReportUpdate(reportType string) {
	// In a production environment, we might want to debounce this to avoid flooding clients
	// if many updates happen in a short burst. For now, we broadcast immediately.

	var data interface{}
	var err error

	// Fetch the latest data based on report type
	// Note: We are fetching "default" views here (e.g., today's data, or standard view).
	// Clients with custom filters might still need to re-fetch manually, but this covers the dashboard/monitoring use case.
	switch reportType {
	case "HOURLY_HEATMAP":
		// Fetch heatmap for the last 7 days by default
		data, err = s.GetHourlySalesHeatmap(time.Now().AddDate(0, 0, -7), time.Now())
	case "SALES_BY_EMPLOYEE":
		// Fetch for today
		data, err = s.GetSalesByEmployeeReport(time.Now().Truncate(24*time.Hour), time.Now())
	case "COGS_GMROI":
		// Fetch for last 30 days
		data, err = s.GetCOGSAndGMROIReport(time.Now().AddDate(0, 0, -30), time.Now())
	case "TAX_LIABILITY":
		// Fetch for current month
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		data, err = s.GetTaxLiabilityReport(startOfMonth, now)
	case "CATEGORY_DRILLDOWN":
		// Fetch for last 30 days
		data, err = s.GetCategoryDrillDownReport(time.Now().AddDate(0, 0, -30), time.Now())
	case "CUSTOMER_INSIGHTS":
		// Fetch for last 90 days
		data, err = s.GetCustomerInsightsReport(time.Now().AddDate(0, 0, -90), time.Now(), true)
	case "SHRINKAGE":
		// Fetch for last 30 days
		data, err = s.GetShrinkageReport(time.Now().AddDate(0, 0, -30), time.Now())
	case "RETURNS_ANALYSIS":
		// Fetch for last 30 days
		data, err = s.GetCustomerReturnAnalysisReport(time.Now().AddDate(0, 0, -30), time.Now())
	default:
		// For unknown or complex reports, we just signal a refetch
		data = map[string]string{"status": "refetch_needed"}
	}

	if err != nil {
		// Log error but don't crash
		// logrus.Errorf("Failed to fetch real-time data for report %s: %v", reportType, err)
		return
	}

	s.hub.BroadcastReportUpdate(reportType, data)
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
		var categoryID, locationID, productID *uint
		if catID, ok := params["categoryId"]; ok && catID != nil {
			val := uint(catID.(float64))
			categoryID = &val
		}
		if locID, ok := params["locationId"]; ok && locID != nil {
			val := uint(locID.(float64))
			locationID = &val
		}
		if prodID, ok := params["productId"]; ok && prodID != nil {
			val := uint(prodID.(float64))
			productID = &val
		}
		groupBy := params["groupBy"].(string)
		err := s.ExportSalesTrendsReport(&buffer, startDate, endDate, categoryID, locationID, productID, groupBy)
		if err != nil {
			return err
		}
	case "sales_trends_pdf":
		fileType = "pdf"
		startDate, _ := time.Parse(time.RFC3339, params["startDate"].(string))
		endDate, _ := time.Parse(time.RFC3339, params["endDate"].(string))
		var categoryID, locationID, productID *uint
		if catID, ok := params["categoryId"]; ok && catID != nil {
			val := uint(catID.(float64))
			categoryID = &val
		}
		if locID, ok := params["locationId"]; ok && locID != nil {
			val := uint(locID.(float64))
			locationID = &val
		}
		if prodID, ok := params["productId"]; ok && prodID != nil {
			val := uint(prodID.(float64))
			productID = &val
		}
		groupBy := params["groupBy"].(string)
		err := s.ExportSalesTrendsReportPDF(&buffer, startDate, endDate, categoryID, locationID, productID, groupBy)
		if err != nil {
			return err
		}
	case "sales_trends_excel":
		fileType = "xlsx"
		startDate, _ := time.Parse(time.RFC3339, params["startDate"].(string))
		endDate, _ := time.Parse(time.RFC3339, params["endDate"].(string))
		var categoryID, locationID, productID *uint
		if catID, ok := params["categoryId"]; ok && catID != nil {
			val := uint(catID.(float64))
			categoryID = &val
		}
		if locID, ok := params["locationId"]; ok && locID != nil {
			val := uint(locID.(float64))
			locationID = &val
		}
		if prodID, ok := params["productId"]; ok && prodID != nil {
			val := uint(prodID.(float64))
			productID = &val
		}
		groupBy := params["groupBy"].(string)
		err := s.ExportSalesTrendsReportExcel(&buffer, startDate, endDate, categoryID, locationID, productID, groupBy)
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

func (s *ReportingService) GetSalesTrendsReport(startDate, endDate time.Time, categoryID, locationID, productID *uint, groupBy string) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("sales_trends:%s:%s:%v:%v:%v:%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), categoryID, locationID, productID, groupBy)
	cachedReport, err := repository.GetCache(cacheKey)
	if err == nil && cachedReport != "" {
		var reportData map[string]interface{}
		err = json.Unmarshal([]byte(cachedReport), &reportData)
		if err == nil {
			return reportData, nil
		}
	}

	salesTrends, topSellingProducts, err := s.repo.GetSalesTrends(startDate, endDate, categoryID, locationID, productID, groupBy)
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

func (s *ReportingService) ExportSalesTrendsReport(writer io.Writer, startDate, endDate time.Time, categoryID, locationID, productID *uint, groupBy string) error {
	reportData, err := s.GetSalesTrendsReport(startDate, endDate, categoryID, locationID, productID, groupBy)
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

func (s *ReportingService) ExportSalesTrendsReportPDF(writer io.Writer, startDate, endDate time.Time, categoryID, locationID, productID *uint, groupBy string) error {
	reportData, err := s.GetSalesTrendsReport(startDate, endDate, categoryID, locationID, productID, groupBy)
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

func (s *ReportingService) ExportSalesTrendsReportExcel(writer io.Writer, startDate, endDate time.Time, categoryID, locationID, productID *uint, groupBy string) error {
	reportData, err := s.GetSalesTrendsReport(startDate, endDate, categoryID, locationID, productID, groupBy)
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

// New Report Wrappers

func (s *ReportingService) GetStockAgingReport() (map[string][]repository.StockAgingItem, error) {
	// Stock aging changes daily, cache for 24h
	cacheKey := "stock_aging_report"
	if cached, err := repository.GetCache(cacheKey); err == nil && cached != "" {
		var reportData map[string][]repository.StockAgingItem
		if err := json.Unmarshal([]byte(cached), &reportData); err == nil {
			return reportData, nil
		}
	}

	data, err := s.repo.GetStockAgingReport()
	if err != nil {
		return nil, err
	}

	if jsonBytes, err := json.Marshal(data); err == nil {
		repository.SetCache(cacheKey, string(jsonBytes), 24*time.Hour)
	}
	return data, nil
}

func (s *ReportingService) GetDeadStockReport(daysThreshold int) ([]repository.DeadStockItem, error) {
	return s.repo.GetDeadStockReport(daysThreshold)
}

func (s *ReportingService) GetSupplierPerformanceReport(startDate, endDate time.Time) ([]repository.SupplierScorecard, error) {
	return s.repo.GetSupplierPerformanceReport(startDate, endDate)
}

func (s *ReportingService) GetHourlySalesHeatmap(startDate, endDate time.Time) ([]repository.HeatmapPoint, error) {
	cacheKey := fmt.Sprintf("heatmap:%s:%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if cached, err := repository.GetCache(cacheKey); err == nil && cached != "" {
		var points []repository.HeatmapPoint
		if err := json.Unmarshal([]byte(cached), &points); err == nil {
			return points, nil
		}
	}

	data, err := s.repo.GetHourlySalesHeatmap(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if jsonBytes, err := json.Marshal(data); err == nil {
		repository.SetCache(cacheKey, string(jsonBytes), 1*time.Hour) // Cache for 1 hour
	}
	return data, nil
}

func (s *ReportingService) GetSalesByEmployeeReport(startDate, endDate time.Time) ([]repository.EmployeeSalesStats, error) {
	return s.repo.GetSalesByEmployeeReport(startDate, endDate)
}

func (s *ReportingService) GetCategoryDrillDownReport(startDate, endDate time.Time) ([]repository.CategoryPerformance, error) {
	return s.repo.GetCategoryDrillDownReport(startDate, endDate)
}

func (s *ReportingService) GetCOGSAndGMROIReport(startDate, endDate time.Time) (repository.GMROIStats, error) {
	return s.repo.GetCOGSAndGMROIReport(startDate, endDate)
}

func (s *ReportingService) GetVoidDiscountAuditReport(startDate, endDate time.Time) ([]domain.AuditLog, error) {
	return s.repo.GetVoidDiscountAuditReport(startDate, endDate)
}

func (s *ReportingService) GetTaxLiabilityReport(startDate, endDate time.Time) ([]repository.TaxLiabilityStats, error) {
	return s.repo.GetTaxLiabilityReport(startDate, endDate)
}

func (s *ReportingService) GetCashDrawerReconciliationReport(startDate, endDate time.Time) ([]domain.CashDrawerSession, error) {
	return s.repo.GetCashDrawerReconciliationReport(startDate, endDate)
}

func (s *ReportingService) GetBasketAnalysisReport(startDate, endDate time.Time, bypassCache bool) ([]repository.BasketAnalysisItem, error) {
	// Basket Analysis is heavy, cache for 24 hours
	cacheKey := fmt.Sprintf("basket_analysis:%s:%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if !bypassCache {
		if cached, err := repository.GetCache(cacheKey); err == nil && cached != "" {
			var items []repository.BasketAnalysisItem
			if err := json.Unmarshal([]byte(cached), &items); err == nil {
				return items, nil
			}
		}
	}

	data, err := s.repo.GetBasketAnalysisReport(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if jsonBytes, err := json.Marshal(data); err == nil {
		repository.SetCache(cacheKey, string(jsonBytes), 24*time.Hour)
	}
	return data, nil
}

func (s *ReportingService) GetPOAnalysisReport(startDate, endDate time.Time) ([]domain.PurchaseOrder, error) {
	return s.repo.GetPOAnalysisReport(startDate, endDate)
}

func (s *ReportingService) GetCustomerInsightsReport(startDate, endDate time.Time, bypassCache bool) ([]repository.CustomerInsight, error) {
	// Customer Insights can be cached for 6 hours
	cacheKey := fmt.Sprintf("customer_insights:%s:%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if !bypassCache {
		if cached, err := repository.GetCache(cacheKey); err == nil && cached != "" {
			var insights []repository.CustomerInsight
			if err := json.Unmarshal([]byte(cached), &insights); err == nil {
				return insights, nil
			}
		}
	}

	data, err := s.repo.GetCustomerInsightsReport(startDate, endDate)
	if err != nil {
		return nil, err
	}

	if jsonBytes, err := json.Marshal(data); err == nil {
		repository.SetCache(cacheKey, string(jsonBytes), 6*time.Hour)
	}
	return data, nil
}

func (s *ReportingService) GetShrinkageReport(startDate, endDate time.Time) ([]repository.ShrinkageItem, error) {
	return s.repo.GetShrinkageReport(startDate, endDate)
}

func (s *ReportingService) GetCustomerReturnAnalysisReport(startDate, endDate time.Time) ([]repository.ReturnAnalysisItem, error) {
	return s.repo.GetCustomerReturnAnalysisReport(startDate, endDate)
}

func (s *ReportingService) GetProductPerformanceAnalytics(startDate, endDate time.Time, supplierName string, minStock int) ([]repository.ProductPerformanceAnalytics, error) {
	return s.repo.GetProductPerformanceAnalytics(startDate, endDate, supplierName, minStock)
}
