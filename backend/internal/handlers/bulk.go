package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/message_broker"
)

// Mock storage for bulk import jobs
var bulkImportJobs = make(map[string]gin.H)

// GetBulkImportJob retrieves a bulk import job from mock storage.
func GetBulkImportJob(jobID string) gin.H {
	job, exists := bulkImportJobs[jobID]
	if !exists {
		return nil
	}
	return job
}

// SetBulkImportJob sets a bulk import job in mock storage.
func SetBulkImportJob(jobID string, job gin.H) {
	bulkImportJobs[jobID] = job
}

// GetProductImportTemplate godoc
// @Summary Download product import template
// @Description Downloads a CSV/Excel template file with required headers for product creation
// @Tags bulk
// @Accept json
// @Produce text/csv
// @Success 200 {file} text/csv "CSV template file"
// @Router /bulk/products/template [get]
func GetProductImportTemplate(c *gin.Context) {
	// In a real application, you would generate a CSV/Excel file dynamically
	// or serve a static template file.
	templateHeaders := "SKU,Name,Description,CategoryID,SubCategoryID,SupplierID,Brand,PurchasePrice,SellingPrice,BarcodeUPC,ImageURLs,Status\n"
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
// @Success 200 {object} map[string]interface{} "Bulk import job status"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/products/import [post]
func UploadProductImport(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get file", http.StatusBadRequest, err))
		return
	}

	// Save the uploaded file temporarily
	uploadDir := "./uploads" // Ensure this directory exists and is writable
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}
	filePath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.Error(appErrors.NewAppError("Failed to save file", http.StatusInternalServerError, err))
		return
	}

	jobID := uuid.New().String()

	// Get UserID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}

	// Publish BulkImportJobEvent
	payload := gin.H{
		"jobId":    jobID,
		"filePath": filePath,
		"userId":   userID.(uint),
	}
	if err := message_broker.Publish("inventory", "bulk.import", payload); err != nil {
		c.Error(appErrors.NewAppError("Failed to publish bulk import event", http.StatusInternalServerError, err))
		return
	}

	// Simulate validation and preview (now handled by a consumer)
	bulkImportJobs[jobID] = gin.H{
		"jobId":          jobID,
		"status":         "QUEUED", // Job is now queued
		"message":        "Bulk import job queued for processing.",
		"filePath":       filePath, // Store path for later confirmation
		"totalRecords":   0,
		"validRecords":   0,
		"invalidRecords": 0,
		"errors":         []gin.H{},
		"preview":        []gin.H{},
	}

	c.JSON(http.StatusOK, bulkImportJobs[jobID])
}

// GetBulkImportStatus godoc
// @Summary Get bulk import job status
// @Description Retrieves the status and validation results of a bulk import job
// @Tags bulk
// @Accept json
// @Produce json
// @Param jobId path string true "Bulk Import Job ID"
// @Success 200 {object} map[string]interface{} "Bulk import job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Router /bulk/products/import/{jobId}/status [get]
func GetBulkImportStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	job, exists := bulkImportJobs[jobID]
	if !exists {
		c.Error(appErrors.NewAppError("Bulk import job not found", http.StatusNotFound, nil))
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
// @Param confirmation body requests.BulkImportConfirmRequest true "Bulk import confirmation request"
// @Success 200 {object} map[string]interface{} "Bulk import confirmation status"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/products/import/{jobId}/confirm [post]
func ConfirmBulkImport(c *gin.Context) {
	jobID := c.Param("jobId")
	job, exists := bulkImportJobs[jobID]
	if !exists {
		c.Error(appErrors.NewAppError("Bulk import job not found", http.StatusNotFound, nil))
		return
	}

	if job["status"] != "PENDING_CONFIRMATION" {
		c.Error(appErrors.NewAppError("Job is not in pending confirmation state", http.StatusBadRequest, nil))
		return
	}

	// Get UserID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}

	// Publish BulkImportJobEvent for confirmation
	payload := gin.H{
		"jobId":    jobID,
		"filePath": job["filePath"].(string),
		"userId":   userID.(uint),
	}
	if err := message_broker.Publish("inventory", "bulk.import", payload); err != nil {
		c.Error(appErrors.NewAppError("Failed to publish bulk import confirmation event", http.StatusInternalServerError, err))
		return
	}

	job["status"] = "PROCESSING"
	bulkImportJobs[jobID] = job
	c.JSON(http.StatusOK, gin.H{"jobId": jobID, "status": "PROCESSING", "message": "Bulk import initiated"})
}

// ExportProducts godoc
// @Summary Export product catalog
// @Description Exports the entire product catalog or a filtered list of products to a CSV/Excel file
// @Tags bulk
// @Accept json
// @Produce text/csv
// @Param format query string false "Export format (csv, excel)" default(csv)
// @Param category query int false "Filter by Category ID"
// @Param supplier query int false "Filter by Supplier ID"
// @Success 200 {file} text/csv "Exported product data"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /bulk/products/export [get]
func ExportProducts(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")

	// Get UserID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}

	jobID := uuid.New().String()
	// Publish BulkExportJobEvent
	payload := gin.H{
		"jobId":    jobID,
		"format":   format,
		"category": c.Query("category"),
		"supplier": c.Query("supplier"),
		"userId":   userID.(uint),
	}
	if err := message_broker.Publish("inventory", "bulk.export", payload); err != nil {
		c.Error(appErrors.NewAppError("Failed to publish bulk export event", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobId": jobID, "status": "QUEUED", "message": "Bulk export job queued for processing."})
}
