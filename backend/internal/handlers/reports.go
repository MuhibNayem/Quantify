package handlers

import (
	"encoding/json"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	appErrors "inventory/backend/internal/errors"
)

type ReportHandler struct {
	reportingService *services.ReportingService
}

func NewReportHandler(reportingService *services.ReportingService) *ReportHandler {
	return &ReportHandler{reportingService: reportingService}
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

	reportType := "sales_trends_" + c.Query("format")
	if reportType == "sales_trends_" {
		reportType = "sales_trends_csv"
	}

	jobID, err := h.reportingService.RequestReport(reportType, req)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to request report", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"jobId": jobID})
}

// GetReportJobStatus godoc
// @Summary Get report job status
// @Description Get the status of a report generation job.
// @Tags reports
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} services.ReportJob "Job status"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/jobs/{jobId} [get]
func (h *ReportHandler) GetReportJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	jobStatus, err := repository.GetCache("report_job:" + jobID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get job status", http.StatusInternalServerError, err))
		return
	}
	if jobStatus == "" {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, nil))
		return
	}

	var job services.ReportJob
	err = json.Unmarshal([]byte(jobStatus), &job)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to parse job status", http.StatusInternalServerError, err))
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
	jobID := c.Param("jobId")
	jobStatus, err := repository.GetCache("report_job:" + jobID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get job status", http.StatusInternalServerError, err))
		return
	}
	if jobStatus == "" {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, nil))
		return
	}

	var job services.ReportJob
	json.Unmarshal([]byte(jobStatus), &job)

	if job.Status != "COMPLETED" {
		c.Error(appErrors.NewAppError("Report not ready", http.StatusNotFound, nil))
		return
	}

	c.Redirect(http.StatusFound, job.FileURL)
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

	logrus.Infof("Generating profit margin report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	reportData, err := h.reportingService.GetProfitMarginReport(req.StartDate, req.EndDate, req.CategoryID, req.LocationID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get profit margin", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, reportData)
}
