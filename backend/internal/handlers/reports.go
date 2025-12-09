package handlers

import (
	"encoding/json"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	appErrors "inventory/backend/internal/errors"
)

type ReportHandler struct {
	reportingService *services.ReportingService
	jobRepo          *repository.JobRepository
}

func NewReportHandler(reportingService *services.ReportingService, jobRepo *repository.JobRepository) *ReportHandler {
	return &ReportHandler{reportingService: reportingService, jobRepo: jobRepo}
}

// GetSalesTrendsReport godoc
// @Summary Get sales trends report
// @Description Generates a report on sales trends over a specified period, with optional filters.
// @Tags reports
// @Accept json
// @Produce json
// @Param request body requests.SalesTrendsReportRequest true "Sales trends report parameters"
// @Success 200 {object} map[string]interface{} "Sales trends data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/sales-trends [post]
func (h *ReportHandler) GetSalesTrendsReport(c *gin.Context) {
	var req requests.SalesTrendsReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if err := validateDateRange(req.StartDate, req.EndDate); err != nil {
		c.Error(err)
		return
	}

	logrus.Infof("Generating sales trends report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	reportData, err := h.reportingService.GetSalesTrendsReport(req.StartDate, req.EndDate, req.CategoryID, req.LocationID, req.GroupBy)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get sales trends", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, reportData)
}

// ExportSalesTrendsReport godoc
// @Summary Export sales trends report
// @Description Exports a sales trends report as a CSV file.
// @Tags reports
// @Accept json
// @Produce text/csv
// @Param request body requests.SalesTrendsReportRequest true "Sales trends report parameters"
// @Success 202 {object} map[string]interface{} "Job ID"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/sales-trends/export [post]
func (h *ReportHandler) ExportSalesTrendsReport(c *gin.Context) {
	var req requests.SalesTrendsReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if err := validateDateRange(req.StartDate, req.EndDate); err != nil {
		c.Error(err)
		return
	}

	reportType := "sales_trends_" + c.Query("format")
	if reportType == "sales_trends_" {
		reportType = "sales_trends_csv"
	}

	payload, err := json.Marshal(req)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to marshal request", http.StatusInternalServerError, err))
		return
	}

	job := &domain.Job{
		Type:       reportType,
		Status:     "QUEUED",
		Payload:    string(payload),
		MaxRetries: 3,
	}

	if err := h.jobRepo.CreateJob(job); err != nil {
		c.Error(appErrors.NewAppError("Failed to create job", http.StatusInternalServerError, err))
		return
	}

	err = message_broker.Publish(c.Request.Context(), "inventory", "reporting", job)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to publish job", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"jobId": job.ID})
}

// GetReportJobStatus godoc
// @Summary Get report job status
// @Description Get the status of a report generation job.
// @Tags reports
// @Produce json
// @Param jobId path int true "Job ID"
// @Success 200 {object} domain.Job "Job status"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/jobs/{jobId} [get]
func (h *ReportHandler) GetReportJobStatus(c *gin.Context) {
	jobID, err := strconv.Atoi(c.Param("jobId"))
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid job ID", http.StatusBadRequest, err))
		return
	}

	job, err := h.jobRepo.GetJob(uint(jobID))
	if err != nil {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, err))
		return
	}

	c.JSON(http.StatusOK, job)
}

// DownloadReportFile godoc
// @Summary Download report file
// @Description Download a generated report file.
// @Tags reports
// @Produce application/octet-stream
// @Param jobId path string true "Job ID"
// @Success 200 {file} file "Report file"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/download/{jobId} [get]
func (h *ReportHandler) DownloadReportFile(c *gin.Context) {
	jobID, err := strconv.Atoi(c.Param("jobId"))
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid job ID", http.StatusBadRequest, err))
		return
	}

	job, err := h.jobRepo.GetJob(uint(jobID))
	if err != nil {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, err))
		return
	}

	if job.Status != "COMPLETED" {
		c.Error(appErrors.NewAppError("Report not ready", http.StatusNotFound, nil))
		return
	}

	c.Redirect(http.StatusFound, job.Result)
}

func (h *ReportHandler) CancelJob(c *gin.Context) {
	jobID, err := strconv.Atoi(c.Param("jobId"))
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid job ID", http.StatusBadRequest, err))
		return
	}

	job, err := h.jobRepo.GetJob(uint(jobID))
	if err != nil {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, err))
		return
	}

	if job.Status == "COMPLETED" || job.Status == "FAILED" {
		c.Error(appErrors.NewAppError("Job already completed or failed", http.StatusBadRequest, nil))
		return
	}

	job.Status = "CANCELLED"
	if err := h.jobRepo.UpdateJob(job); err != nil {
		c.Error(appErrors.NewAppError("Failed to cancel job", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job cancelled successfully"})
}

func validateDateRange(start, end time.Time) *appErrors.AppError {
	if end.Before(start) {
		return appErrors.NewAppError("End date must be after start date", http.StatusBadRequest, nil)
	}
	return nil
}

// GetInventoryTurnoverReport godoc
// @Summary Get inventory turnover report
// @Description Generates a report on inventory turnover rate over a specified period.
// @Tags reports
// @Accept json
// @Produce json
// @Param request body requests.InventoryTurnoverReportRequest true "Inventory turnover report parameters"
// @Success 200 {object} map[string]interface{} "Inventory turnover data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/inventory-turnover [post]
func (h *ReportHandler) GetInventoryTurnoverReport(c *gin.Context) {
	var req requests.InventoryTurnoverReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if err := validateDateRange(req.StartDate, req.EndDate); err != nil {
		c.Error(err)
		return
	}

	logrus.Infof("Generating inventory turnover report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	reportData, err := h.reportingService.GetInventoryTurnoverReport(req.StartDate, req.EndDate, req.CategoryID, req.LocationID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get inventory turnover", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, reportData)
}

// GetProfitMarginReport godoc
// @Summary Get profit margin report
// @Description Generates a report on profit margins for products or categories over a specified period.
// @Tags reports
// @Accept json
// @Produce json
// @Param request body requests.ProfitMarginReportRequest true "Profit margin report parameters"
// @Success 200 {object} map[string]interface{} "Profit margin data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/profit-margin [post]
func (h *ReportHandler) GetProfitMarginReport(c *gin.Context) {
	var req requests.ProfitMarginReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if err := validateDateRange(req.StartDate, req.EndDate); err != nil {
		c.Error(err)
		return
	}

	logrus.Infof("Generating profit margin report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	reportData, err := h.reportingService.GetProfitMarginReport(req.StartDate, req.EndDate, req.CategoryID, req.LocationID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get profit margin", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, reportData)
}

// GetStockAgingReport godoc
// @Summary Get stock aging report
// @Description Groups inventory by age buckets.
// @Tags reports
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /reports/stock-aging [get]
func (h *ReportHandler) GetStockAgingReport(c *gin.Context) {
	report, err := h.reportingService.GetStockAgingReport()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get stock aging report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetDeadStockReport godoc
// @Summary Get dead stock report
// @Description Identifies products with no sales in X days.
// @Tags reports
// @Produce json
// @Param days query int false "Days threshold (default 90)"
// @Success 200 {array} repository.DeadStockItem
// @Router /reports/dead-stock [get]
func (h *ReportHandler) GetDeadStockReport(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "90")
	days, _ := strconv.Atoi(daysStr)

	report, err := h.reportingService.GetDeadStockReport(days)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get dead stock report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetSupplierPerformanceReport godoc
// @Summary Get supplier performance scorecard
// @Description Metrics for suppliers.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.SupplierScorecard
// @Router /reports/supplier-performance [get]
func (h *ReportHandler) GetSupplierPerformanceReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetSupplierPerformanceReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get supplier performance report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetHourlySalesHeatmap godoc
// @Summary Get hourly sales heatmap
// @Description Sales by hour and day of week.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.HeatmapPoint
// @Router /reports/heatmap [get]
func (h *ReportHandler) GetHourlySalesHeatmap(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetHourlySalesHeatmap(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get heatmap", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetSalesByEmployeeReport godoc
// @Summary Get sales by employee
// @Description Sales totals by user.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.EmployeeSalesStats
// @Router /reports/employee-sales [get]
func (h *ReportHandler) GetSalesByEmployeeReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetSalesByEmployeeReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get employee sales report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetCategoryDrillDownReport godoc
// @Summary Get category drill-down
// @Description Sales and margin by category.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.CategoryPerformance
// @Router /reports/category-drilldown [get]
func (h *ReportHandler) GetCategoryDrillDownReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetCategoryDrillDownReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get category report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetCOGSAndGMROIReport godoc
// @Summary Get COGS and GMROI
// @Description GMROI analysis.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {object} repository.GMROIStats
// @Router /reports/gmroi [get]
func (h *ReportHandler) GetCOGSAndGMROIReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetCOGSAndGMROIReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get GMROI report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetVoidDiscountAuditReport godoc
// @Summary Get void/discount audit log
// @Description Audit log for sensitive POS actions.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} domain.AuditLog
// @Router /reports/audit/voids [get]
func (h *ReportHandler) GetVoidDiscountAuditReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetVoidDiscountAuditReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get audit report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetTaxLiabilityReport godoc
// @Summary Get tax liability report
// @Description Tax collected.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.TaxLiabilityStats
// @Router /reports/tax-liability [get]
func (h *ReportHandler) GetTaxLiabilityReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetTaxLiabilityReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get tax report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetCashDrawerReconciliationReport godoc
// @Summary Get cash drawer reconciliation
// @Description Cash drawer sessions.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} domain.CashDrawerSession
// @Router /reports/cash-reconciliation [get]
func (h *ReportHandler) GetCashDrawerReconciliationReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetCashDrawerReconciliationReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get cash reconciliation report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetCustomerInsightsReport godoc
// @Summary Get customer insights
// @Description Top spenders and churn risk.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.CustomerInsight
// @Router /reports/customer-insights [get]
func (h *ReportHandler) GetCustomerInsightsReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetCustomerInsightsReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get customer insights", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetShrinkageReport godoc
// @Summary Get shrinkage report
// @Description Inventory loss analysis.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.ShrinkageItem
// @Router /reports/shrinkage [get]
func (h *ReportHandler) GetShrinkageReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetShrinkageReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get shrinkage report", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetCustomerReturnAnalysisReport godoc
// @Summary Get returns analysis
// @Description Analysis of customer returns.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.ReturnAnalysisItem
// @Router /reports/returns-analysis [get]
func (h *ReportHandler) GetCustomerReturnAnalysisReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetCustomerReturnAnalysisReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get returns analysis", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetBasketAnalysisReport godoc
// @Summary Get basket analysis (frequently bought together)
// @Description Identifies product pairs often bought in the same order.
// @Tags reports
// @Produce json
// @Param startDate query string true "Start Date (RFC3339)"
// @Param endDate query string true "End Date (RFC3339)"
// @Success 200 {array} repository.BasketAnalysisItem
// @Router /reports/basket-analysis [get]
func (h *ReportHandler) GetBasketAnalysisReport(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		c.Error(err)
		return
	}

	report, err := h.reportingService.GetBasketAnalysisReport(start, end)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get basket analysis", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, report)
}

// Helper to parse start/end date query params
func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.Query("startDate")
	endStr := c.Query("endDate")

	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, appErrors.NewAppError("startDate and endDate are required", http.StatusBadRequest, nil)
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return time.Time{}, time.Time{}, appErrors.NewAppError("Invalid startDate format (use RFC3339)", http.StatusBadRequest, err)
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return time.Time{}, time.Time{}, appErrors.NewAppError("Invalid endDate format (use RFC3339)", http.StatusBadRequest, err)
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, appErrors.NewAppError("End date must be after start date", http.StatusBadRequest, nil)
	}

	return start, end, nil
}
