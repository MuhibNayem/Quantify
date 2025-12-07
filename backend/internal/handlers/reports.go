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
