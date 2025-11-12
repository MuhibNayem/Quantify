package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
	"inventory/backend/internal/storage"
	ws "inventory/backend/internal/websocket"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BulkConsumer struct {
	DB *gorm.DB

	JobRepo *repository.JobRepository

	ProductRepo *repository.ProductRepository

	CategoryRepo *repository.CategoryRepository

	SupplierRepo *repository.SupplierRepository

	Uploader storage.Uploader

	BulkImportSvc *services.BulkImportService

	BulkExportSvc *services.BulkExportService

	Hub *ws.Hub
}

func NewBulkConsumer(db *gorm.DB, jobRepo *repository.JobRepository, productRepo *repository.ProductRepository, categoryRepo *repository.CategoryRepository, supplierRepo *repository.SupplierRepository, uploader storage.Uploader, bulkImportSvc *services.BulkImportService, bulkExportSvc *services.BulkExportService, hub *ws.Hub) *BulkConsumer {

	return &BulkConsumer{

		DB: db,

		JobRepo: jobRepo,

		ProductRepo: productRepo,

		CategoryRepo: categoryRepo,

		SupplierRepo: supplierRepo,

		Uploader: uploader,

		BulkImportSvc: bulkImportSvc,

		BulkExportSvc: bulkExportSvc,

		Hub: hub,
	}

}

func (c *BulkConsumer) Start(ctx context.Context, workers int) context.CancelFunc {
	cancelFuncs := make([]context.CancelFunc, 0, workers*3)
	for i := 0; i < workers; i++ {
		cancelFuncs = append(cancelFuncs, message_broker.Subscribe(ctx, "inventory", "bulk-import-queue", "bulk.import", c.handleImport))
		cancelFuncs = append(cancelFuncs, message_broker.Subscribe(ctx, "inventory", "bulk-import-confirm-queue", "bulk.import.confirm", c.handleImportConfirm))
		cancelFuncs = append(cancelFuncs, message_broker.Subscribe(ctx, "inventory", "bulk-export-queue", "bulk.export", c.handleExport))
	}
	logrus.Infof("Bulk consumer started with %d worker(s)", workers)

	return func() {
		for _, cancel := range cancelFuncs {
			cancel()
		}
	}
}

func (c *BulkConsumer) handleImport(ctx context.Context, deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.processImportDelivery(d)
		}
	}
}

func (c *BulkConsumer) processImportDelivery(d amqp091.Delivery) {
	var msgPayload map[string]interface{}
	if err := json.Unmarshal(d.Body, &msgPayload); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		d.Nack(false, false)
		return
	}

	jobIDFloat, ok := msgPayload["jobId"].(float64)
	if !ok {
		logrus.Errorf("Invalid job ID in message payload")
		return
	}
	jobID := uint(jobIDFloat)

	job, err := c.JobRepo.GetJob(jobID)
	if err != nil {
		logrus.Errorf("Failed to get job %d: %v", jobID, err)
		d.Nack(false, false)
		return
	}

	job.Status = "PROCESSING"
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to update job %d to PROCESSING: %v", jobID, err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		logrus.Errorf("Failed to unmarshal job payload: %v", err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after payload error: %v", jobID, err)
		}
		d.Nack(false, false)
		return
	}

	bucketName, ok := payload["bucketName"].(string)
	if !ok || bucketName == "" {
		logrus.Errorf("Bucket name missing in payload for job %d", jobID)
		job.Status = "FAILED"
		job.LastError = "bucket name missing in payload"
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after payload error: %v", jobID, err)
		}
		d.Nack(false, false)
		return
	}

	objectName, ok := payload["objectName"].(string)
	if !ok || objectName == "" {
		logrus.Errorf("Object name missing in payload for job %d", jobID)
		job.Status = "FAILED"
		job.LastError = "object name missing in payload"
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after payload error: %v", jobID, err)
		}
		d.Nack(false, false)
		return
	}

	file, err := c.Uploader.DownloadFile(bucketName, objectName)
	if err != nil {
		logrus.Errorf("Failed to download file %s from bucket %s: %v", objectName, bucketName, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after download error: %v", jobID, err)
		}
		d.Nack(false, false)
		return
	}
	defer file.Close()

	result, err := c.BulkImportSvc.ProcessBulkImport(file)
	if err != nil {
		logrus.Errorf("Failed to process bulk import for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after validation error: %v", jobID, err)
		}
		d.Nack(false, false)
		return
	}

	resultBytes, _ := json.Marshal(result)
	job.Result = string(resultBytes)
	if result.ValidRecords == 0 {
		job.Status = "FAILED"
		job.LastError = "No valid products found in uploaded file"
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after validation: %v", jobID, err)
		}
		logrus.Warnf("Bulk import job %d finished validation with no valid records", jobID)
		d.Ack(false)
		return
	}

	job.Status = "PENDING_CONFIRMATION"
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to update job %d: %v", jobID, err)
		d.Nack(false, false)
		return
	}

	if err := c.finalizeImport(job, result); err != nil {
		logrus.Errorf("Failed to finalize bulk import job %d: %v", jobID, err)
		d.Nack(false, false)
		return
	}

	logrus.Infof("Bulk import job %d completed successfully after auto-confirmation", jobID)
	d.Ack(false)
}

func (c *BulkConsumer) handleImportConfirm(ctx context.Context, deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.processImportConfirmDelivery(d)
		}
	}
}

func (c *BulkConsumer) processImportConfirmDelivery(d amqp091.Delivery) {

	var msgPayload map[string]interface{}

	if err := json.Unmarshal(d.Body, &msgPayload); err != nil {

		logrus.Errorf("Failed to unmarshal message payload: %v", err)

		d.Nack(false, false)

		return

	}

	jobIDFloat, ok := msgPayload["jobId"].(float64)

	if !ok {

		logrus.Errorf("Invalid job ID in message payload")

		d.Nack(false, false)

		return

	}

	jobID := uint(jobIDFloat)

	job, err := c.JobRepo.GetJob(jobID)

	if err != nil {

		logrus.Errorf("Failed to get job %d: %v", jobID, err)

		d.Nack(true, true) // Requeue

		return

	}

	if job.Status != "PENDING_CONFIRMATION" {

		logrus.Warnf("Job %d is not in PENDING_CONFIRMATION state, skipping confirmation", jobID)

		d.Ack(false)

		return

	}

	job.Status = "PROCESSING"
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to update job %d to PROCESSING: %v", jobID, err)
		d.Nack(false, false)
		return
	}

	var importResult services.BulkImportResult

	if err := json.Unmarshal([]byte(job.Result), &importResult); err != nil {

		logrus.Errorf("Failed to unmarshal import result for job %d: %v", jobID, err)

		job.Status = "FAILED"

		job.LastError = err.Error()

		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after result error: %v", jobID, err)
		}

		d.Nack(false, false)

		return

	}

	if err := c.finalizeImport(job, &importResult); err != nil {
		logrus.Errorf("Failed to finalize import for job %d: %v", jobID, err)
		d.Nack(false, false)
		return
	}

	d.Ack(false)
}

func (c *BulkConsumer) saveJob(job *domain.Job) error {
	if err := c.JobRepo.UpdateJob(job); err != nil {
		return err
	}
	c.emitJobUpdate(job)
	return nil
}

func (c *BulkConsumer) emitJobUpdate(job *domain.Job) {
	if c.Hub == nil {
		return
	}

	userID := c.extractJobUserID(job)
	payload := gin.H{
		"event":     "BULK_JOB_STATUS",
		"jobId":     job.ID,
		"type":      job.Type,
		"status":    job.Status,
		"result":    job.Result,
		"lastError": job.LastError,
	}

	if userID > 0 {
		c.Hub.SendToUser(userID, payload)
	} else {
		c.Hub.Broadcast(payload)
	}
}

func (c *BulkConsumer) extractJobUserID(job *domain.Job) uint {
	if job.Payload == "" {
		return 0
	}
	var payload struct {
		UserID uint `json:"userId"`
	}
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		logrus.Errorf("Failed to parse job payload for job %d: %v", job.ID, err)
		return 0
	}
	return payload.UserID
}

func (c *BulkConsumer) finalizeImport(job *domain.Job, importResult *services.BulkImportResult) error {
	tx := c.DB.Begin()
	if tx.Error != nil {
		c.failJob(job, "failed to begin transaction for import finalization")
		return tx.Error
	}

	normalize := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}

	categoryNameIDMap := make(map[string]uint)
	var existingCategories []domain.Category
	if err := tx.Find(&existingCategories).Error; err != nil {
		tx.Rollback()
		c.failJob(job, "failed to load categories during import finalization")
		return err
	}
	for _, cat := range existingCategories {
		categoryNameIDMap[normalize(cat.Name)] = cat.ID
	}

	supplierNameIDMap := make(map[string]uint)
	var existingSuppliers []domain.Supplier
	if err := tx.Find(&existingSuppliers).Error; err != nil {
		tx.Rollback()
		c.failJob(job, "failed to load suppliers during import finalization")
		return err
	}
	for _, sup := range existingSuppliers {
		supplierNameIDMap[normalize(sup.Name)] = sup.ID
	}

	subCategoryKey := func(categoryID uint, name string) string {
		return fmt.Sprintf("%d::%s", categoryID, normalize(name))
	}
	subCategoryKeyIDMap := make(map[string]uint)
	var existingSubCategories []domain.SubCategory
	if err := tx.Find(&existingSubCategories).Error; err != nil {
		tx.Rollback()
		c.failJob(job, "failed to load sub-categories during import finalization")
		return err
	}
	for _, subCat := range existingSubCategories {
		subCategoryKeyIDMap[subCategoryKey(subCat.CategoryID, subCat.Name)] = subCat.ID
	}

	locationNameIDMap := make(map[string]uint)
	var existingLocations []domain.Location
	if err := tx.Find(&existingLocations).Error; err != nil {
		tx.Rollback()
		c.failJob(job, "failed to load locations during import finalization")
		return err
	}
	for _, loc := range existingLocations {
		locationNameIDMap[normalize(loc.Name)] = loc.ID
	}

	createSupplier := func(name string) error {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return nil
		}
		key := normalize(trimmed)
		if _, exists := supplierNameIDMap[key]; exists {
			return nil
		}
		supplier := &domain.Supplier{Name: trimmed}
		if err := tx.Create(supplier).Error; err != nil {
			return err
		}
		supplierNameIDMap[key] = supplier.ID
		return nil
	}

	createCategory := func(name string) (uint, error) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return 0, fmt.Errorf("category name is required for sub-category creation")
		}
		key := normalize(trimmed)
		if id, exists := categoryNameIDMap[key]; exists {
			return id, nil
		}
		category := &domain.Category{Name: trimmed}
		if err := tx.Create(category).Error; err != nil {
			return 0, err
		}
		categoryNameIDMap[key] = category.ID
		return category.ID, nil
	}

	createSubCategory := func(name string, categoryID uint) (uint, error) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" || categoryID == 0 {
			return 0, nil
		}
		key := subCategoryKey(categoryID, trimmed)
		if id, exists := subCategoryKeyIDMap[key]; exists {
			return id, nil
		}
		subCat := &domain.SubCategory{Name: trimmed, CategoryID: categoryID}
		if err := tx.Create(subCat).Error; err != nil {
			return 0, err
		}
		subCategoryKeyIDMap[key] = subCat.ID
		return subCat.ID, nil
	}

	createLocation := func(name string) (uint, error) {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return 0, fmt.Errorf("location name is required")
		}
		key := normalize(trimmed)
		if id, exists := locationNameIDMap[key]; exists {
			return id, nil
		}
		location := &domain.Location{Name: trimmed}
		if err := tx.Create(location).Error; err != nil {
			return 0, err
		}
		locationNameIDMap[key] = location.ID
		return location.ID, nil
	}

	for i := range importResult.ValidProducts {
		p := &importResult.ValidProducts[i]
		locName := strings.TrimSpace(p.Location.Name)
		locID, err := createLocation(locName)
		if err != nil {
			tx.Rollback()
			c.failJob(job, fmt.Sprintf("failed to ensure location '%s'", locName))
			return err
		}
		if err := createSupplier(p.Supplier.Name); err != nil {
			tx.Rollback()
			c.failJob(job, fmt.Sprintf("failed to ensure supplier '%s'", p.Supplier.Name))
			return err
		}

		var categoryID uint
		catName := strings.TrimSpace(p.Category.Name)
		if catName != "" {
			id, err := createCategory(catName)
			if err != nil {
				tx.Rollback()
				c.failJob(job, fmt.Sprintf("failed to ensure category '%s'", catName))
				return err
			}
			categoryID = id
		}

		var subCategoryID uint
		subName := strings.TrimSpace(p.SubCategory.Name)
		if subName != "" {
			if categoryID == 0 {
				logrus.Warnf("Row missing category for sub-category '%s'; skipping sub-category creation", subName)
			} else {
				id, err := createSubCategory(subName, categoryID)
				if err != nil {
					tx.Rollback()
					c.failJob(job, fmt.Sprintf("failed to create sub-category '%s'", subName))
					return err
				}
				subCategoryID = id
			}
		}

		if catName != "" {
			categoryKey := normalize(catName)
			if id, ok := categoryNameIDMap[categoryKey]; ok {
				categoryID = id
			}
		}

		if categoryID == 0 && catName != "" {
			logrus.Warnf("Category '%s' could not be resolved for job %d", catName, job.ID)
		}

		if subName != "" && categoryID == 0 {
			logrus.Warnf("Skipping sub-category '%s' due to missing parent category for job %d", subName, job.ID)
		}

		if supplierName := strings.TrimSpace(p.Supplier.Name); supplierName != "" {
			if id, ok := supplierNameIDMap[normalize(supplierName)]; ok {
				p.SupplierID = id
			}
		}

		p.CategoryID = categoryID
		p.SubCategoryID = subCategoryID
		p.LocationID = locID

		// Prevent GORM from trying to upsert nested entities during CreateInBatches.
		p.Category = domain.Category{}
		p.SubCategory = domain.SubCategory{}
		p.Supplier = domain.Supplier{}
		p.Location = domain.Location{}
	}

	if err := tx.CreateInBatches(importResult.ValidProducts, 100).Error; err != nil {
		tx.Rollback()
		c.failJob(job, fmt.Sprintf("failed to bulk insert products: %v", err))
		return err
	}

	if err := tx.Commit().Error; err != nil {
		c.failJob(job, "failed to commit transaction")
		return err
	}

	job.Status = "COMPLETED"
	job.LastError = ""
	return c.saveJob(job)
}

func (c *BulkConsumer) failJob(job *domain.Job, lastError string) {
	job.Status = "FAILED"
	job.LastError = lastError
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to persist failure state for job %d: %v", job.ID, err)
	}
}

func (c *BulkConsumer) handleExport(ctx context.Context, deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.processExportDelivery(d)
		}
	}
}

func (c *BulkConsumer) processExportDelivery(d amqp091.Delivery) {
	var msgPayload map[string]interface{}
	if err := json.Unmarshal(d.Body, &msgPayload); err != nil {
		logrus.Errorf("Failed to unmarshal message payload: %v", err)
		return
	}

	jobIDFloat, ok := msgPayload["jobId"].(float64)
	if !ok {
		logrus.Errorf("Invalid job ID in message payload")
		return
	}
	jobID := uint(jobIDFloat)

	job, err := c.JobRepo.GetJob(jobID)
	if err != nil {
		logrus.Errorf("Failed to get job %d: %v", jobID, err)
		return
	}

	job.Status = "PROCESSING"
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to update export job %d to PROCESSING: %v", jobID, err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		logrus.Errorf("Failed to unmarshal job payload: %v", err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after payload error: %v", jobID, err)
		}
		return
	}

	// Fetch products from the database
	// This is a simplified example. In a real application, you would use filters from the payload.
	products, err := c.ProductRepo.GetAll()
	if err != nil {
		logrus.Errorf("Failed to get products for export job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after product fetch error: %v", jobID, err)
		}
		return
	}

	format := payload["format"].(string)
	buffer, err := c.BulkExportSvc.GenerateProductExport(products, format)
	if err != nil {
		logrus.Errorf("Failed to generate export file for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after export generation error: %v", jobID, err)
		}
		return
	}

	bucketName := "bulk-exports"
	objectName := fmt.Sprintf("%d-export.%s", jobID, format)
	_, err = c.Uploader.UploadFile(bucketName, objectName, buffer, int64(buffer.Len()))
	if err != nil {
		logrus.Errorf("Failed to upload export file for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after upload error: %v", jobID, err)
		}
		return
	}

	result := gin.H{
		"downloadUrl": fmt.Sprintf("/files/%s/%s", bucketName, objectName),
	}
	resultBytes, _ := json.Marshal(result)
	job.Result = string(resultBytes)
	job.Status = "COMPLETED"
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to update job %d: %v", jobID, err)
	}
}
