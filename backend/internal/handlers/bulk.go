package handlers

import (
	"encoding/json"
	"fmt"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/storage"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BulkHandler struct {
	JobRepo  *repository.JobRepository
	Uploader storage.Uploader
}

func NewBulkHandler(jobRepo *repository.JobRepository, uploader storage.Uploader) *BulkHandler {
	return &BulkHandler{
		JobRepo:  jobRepo,
		Uploader: uploader,
	}
}

// GetProductImportTemplate godoc
// @Summary Download product import template
// @Description Downloads a CSV/Excel template file with required headers for product creation
// @Tags bulk
// @Accept json
// @Produce text/csv
// @Success 200 {file} text/csv "CSV template file"
// @Router /bulk/products/template [get]
func (h *BulkHandler) GetProductImportTemplate(c *gin.Context) {
	templateHeaders := "SKU,Name,Description,CategoryName,SubCategoryName,SupplierName,Brand,PurchasePrice,SellingPrice,LocationName,Status\n"
	c.Header("Content-Disposition", "attachment; filename=product_import_template.csv")
	c.Data(http.StatusOK, "text/csv", []byte(templateHeaders))
}

// UploadProductImport godoc
// @Summary Upload a file for bulk product import
// @Description Uploads a CSV/Excel file for bulk product creation/update. Returns a job ID for status tracking.
// @Tags bulk
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV/Excel file to upload"
// @Success 202 {object} domain.Job "Bulk import job started"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/products/import [post]
func (h *BulkHandler) UploadProductImport(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get file", http.StatusBadRequest, err))
		return
	}

	// Get UserID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}

	bucketName := "bulk-imports"
	objectName := fmt.Sprintf("%s-%s", uuid.New().String(), file.Filename)

	_, err = h.Uploader.UploadFileFromMultipart(bucketName, objectName, file)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to upload file", http.StatusInternalServerError, err))
		return
	}

	payload := gin.H{
		"bucketName": bucketName,
		"objectName": objectName,
		"userId":     userID.(uint),
	}
	payloadBytes, _ := json.Marshal(payload)

	job := &domain.Job{
		Type:       "BULK_IMPORT",
		Status:     "QUEUED",
		Payload:    string(payloadBytes),
		MaxRetries: 3,
	}

	if err := h.JobRepo.CreateJob(job); err != nil {
		c.Error(appErrors.NewAppError("Failed to create job", http.StatusInternalServerError, err))
		return
	}

	// Publish message to RabbitMQ
	msgPayload := gin.H{"jobId": job.ID}
	if err := message_broker.Publish(c.Request.Context(), "inventory", "bulk.import", msgPayload); err != nil {
		c.Error(appErrors.NewAppError("Failed to publish bulk import event", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusAccepted, job)
}

// GetBulkImportStatus godoc
// @Summary Get bulk import job status
// @Description Retrieves the status and validation results of a bulk import job
// @Tags bulk
// @Accept json
// @Produce json
// @Param jobId path string true "Bulk Import Job ID"
// @Success 200 {object} domain.Job "Bulk import job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /bulk/products/import/{jobId}/status [get]
func (h *BulkHandler) GetBulkImportStatus(c *gin.Context) {
	jobIDStr := c.Param("jobId")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid job ID", http.StatusBadRequest, err))
		return
	}

	job, err := h.JobRepo.GetJob(uint(jobID))
	if err != nil {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, err))
		return
	}

	c.JSON(http.StatusOK, job)
}

// ConfirmBulkImport godoc
// @Summary Confirm and execute bulk import
// @Description Confirms and executes the bulk import after preview
// @Tags bulk
// @Accept json
// @Produce json
// @Param jobId path string true "Bulk Import Job ID"
// @Success 202 {object} map[string]interface{} "Bulk import confirmation status"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/products/import/{jobId}/confirm [post]
func (h *BulkHandler) ConfirmBulkImport(c *gin.Context) {
	jobIDStr := c.Param("jobId")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid job ID", http.StatusBadRequest, err))
		return
	}

	job, err := h.JobRepo.GetJob(uint(jobID))
	if err != nil {
		c.Error(appErrors.NewAppError("Job not found", http.StatusNotFound, err))
		return
	}

	if job.Status != "PENDING_CONFIRMATION" {
		c.Error(appErrors.NewAppError("Job is not in pending confirmation state", http.StatusBadRequest, nil))
		return
	}

	// Publish confirmation event
	msgPayload := gin.H{"jobId": job.ID}
	if err := message_broker.Publish(c.Request.Context(), "inventory", "bulk.import.confirm", msgPayload); err != nil {
		c.Error(appErrors.NewAppError("Failed to publish bulk import confirmation event", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"jobId": job.ID, "status": "PENDING_CONFIRMATION", "message": "Bulk import confirmation received"})
}

// ExportProducts godoc
// @Summary Export product catalog
// @Description Exports the entire product catalog or a filtered list of products to a CSV/Excel file
// @Tags bulk
// @Accept json
// @Produce json
// @Param format query string false "Export format (csv, excel)" default(csv)
// @Param category query int false "Filter by Category ID"
// @Param supplier query int false "Filter by Supplier ID"
// @Success 202 {object} domain.Job "Export job started"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/products/export [get]
func (h *BulkHandler) ExportProducts(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")

	// Get UserID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}

	payload := gin.H{
		"format":   format,
		"category": c.Query("category"),
		"supplier": c.Query("supplier"),
		"userId":   userID.(uint),
	}
	payloadBytes, _ := json.Marshal(payload)

	job := &domain.Job{
		Type:       "BULK_EXPORT",
		Status:     "QUEUED",
		Payload:    string(payloadBytes),
		MaxRetries: 3,
	}

	if err := h.JobRepo.CreateJob(job); err != nil {
		c.Error(appErrors.NewAppError("Failed to create job", http.StatusInternalServerError, err))
		return
	}

	// Publish message to RabbitMQ
	msgPayload := gin.H{"jobId": job.ID}
	if err := message_broker.Publish(c.Request.Context(), "inventory", "bulk.export", msgPayload); err != nil {
		c.Error(appErrors.NewAppError("Failed to publish bulk export event", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusAccepted, job)
}

// ListBulkJobs godoc
// @Summary List recent bulk jobs
// @Description Retrieves a list of the most recent bulk import/export jobs
// @Tags bulk
// @Accept json
// @Produce json
// @Success 200 {array} domain.Job "List of recent jobs"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/jobs [get]
func (h *BulkHandler) ListBulkJobs(c *gin.Context) {
	jobs, err := h.JobRepo.ListJobs(50)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to list jobs", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// DownloadFile godoc
// @Summary Download a file from bulk storage
// @Description Downloads a file (import/export) from the bulk storage bucket
// @Tags bulk
// @Accept json
// @Produce octet-stream
// @Param bucket path string true "Bucket Name"
// @Param object path string true "Object Name"
// @Success 200 {file} application/octet-stream "File content"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/files/{bucket}/{object} [get]
func (h *BulkHandler) DownloadFile(c *gin.Context) {
	bucket := c.Param("bucket")
	object := c.Param("object")

	if bucket == "" || object == "" {
		c.Error(appErrors.NewAppError("Bucket and object names are required", http.StatusBadRequest, nil))
		return
	}

	// Security check: Ensure bucket is one of the allowed bulk buckets
	if bucket != "bulk-imports" && bucket != "bulk-exports" {
		c.Error(appErrors.NewAppError("Invalid bucket access", http.StatusForbidden, nil))
		return
	}

	file, err := h.Uploader.DownloadFile(bucket, object)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to download file", http.StatusNotFound, err))
		return
	}
	defer file.Close()

	// Set headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", object))
	c.Header("Content-Type", "application/octet-stream")

	// Stream the file to the response
	// Note: io.Copy is efficient for streaming
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		// If streaming fails mid-way, we can't really change the status code, but we can log it
		// c.Error is still useful for logging
		c.Error(appErrors.NewAppError("Failed to stream file", http.StatusInternalServerError, err))
	}
}
