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
	"time"

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

	LocationRepo *repository.LocationRepository

	Uploader storage.Uploader

	BulkImportService *services.BulkImportService

	BulkExportSvc *services.BulkExportService

	Hub *ws.Hub

	NotificationRepo repository.NotificationRepository
}

func NewBulkConsumer(db *gorm.DB, jobRepo *repository.JobRepository, productRepo *repository.ProductRepository, categoryRepo *repository.CategoryRepository, supplierRepo *repository.SupplierRepository, locationRepo *repository.LocationRepository, uploader storage.Uploader, bulkImportSvc *services.BulkImportService, bulkExportSvc *services.BulkExportService, hub *ws.Hub, notificationRepo repository.NotificationRepository) *BulkConsumer {

	return &BulkConsumer{

		DB: db,

		JobRepo: jobRepo,

		ProductRepo: productRepo,

		CategoryRepo: categoryRepo,

		SupplierRepo: supplierRepo,

		LocationRepo: locationRepo,

		Uploader: uploader,

		BulkImportService: bulkImportSvc,

		BulkExportSvc: bulkExportSvc,

		Hub: hub,

		NotificationRepo: notificationRepo,
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

	result, err := c.BulkImportService.ProcessBulkImport(file)
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
		// Persist for user
		userNotif := domain.Notification{
			UserID:      userID,
			Type:        "BULK_JOB_STATUS",
			Title:       fmt.Sprintf("Bulk Job %s", job.Status),
			Message:     fmt.Sprintf("Job #%d (%s) is now %s.", job.ID, job.Type, job.Status),
			TriggeredAt: time.Now(),
		}
		c.NotificationRepo.CreateNotification(&userNotif)
	} else {
		// Fallback: Broadcast only to admins/managers with bulk import permission
		c.Hub.BroadcastToPermission("bulk.import", payload)
		// Persist for admins
		adminNotif := domain.Notification{
			Type:        "BULK_JOB_STATUS",
			Title:       fmt.Sprintf("System Bulk Job %s", job.Status),
			Message:     fmt.Sprintf("System Job #%d (%s) is now %s.", job.ID, job.Type, job.Status),
			TriggeredAt: time.Now(),
		}
		c.NotificationRepo.CreateNotificationsForPermission("bulk.import", adminNotif)
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

	// Maps to store IDs of newly created entities
	newCategoryIDs := make(map[string]uint)
	newSupplierIDs := make(map[string]uint)
	newLocationIDs := make(map[string]uint)

	// 1. Create New Categories
	for name := range importResult.NewEntities.Categories {
		cat := &domain.Category{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(cat).Error; err != nil {
			tx.Rollback()
			c.failJob(job, fmt.Sprintf("failed to create category '%s': %v", name, err))
			return err
		}
		newCategoryIDs[name] = cat.ID
	}

	// 2. Create New Suppliers
	for name := range importResult.NewEntities.Suppliers {
		sup := &domain.Supplier{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(sup).Error; err != nil {
			tx.Rollback()
			c.failJob(job, fmt.Sprintf("failed to create supplier '%s': %v", name, err))
			return err
		}
		newSupplierIDs[name] = sup.ID
	}

	// 3. Create New Locations
	for name := range importResult.NewEntities.Locations {
		loc := &domain.Location{Name: name}
		if err := tx.Where("name = ?", name).FirstOrCreate(loc).Error; err != nil {
			tx.Rollback()
			c.failJob(job, fmt.Sprintf("failed to create location '%s': %v", name, err))
			return err
		}
		newLocationIDs[name] = loc.ID
	}

	// 4. Create New SubCategories
	// We need to iterate ValidProducts to find the parent CategoryID for each new SubCategory
	// OR we can rely on the fact that ProcessBulkImport validated that parent exists or is being created.
	// But NewEntities.SubCategories is just a set of names. We don't know the parent from that set.
	// We need to look at ValidProducts to find a product that uses this SubCategory and get its CategoryID.
	// This is a bit tricky if multiple products use the same new SubCategory but with different Categories (which shouldn't happen for the same SubCategory name usually, but names are not unique across categories in some systems. In ours, SubCategory name is unique per Category).
	// Let's iterate ValidProducts to create SubCategories on demand if they are new.

	createdSubCategories := make(map[string]uint) // Key: "CategoryID::SubName"

	for i := range importResult.ValidProducts {
		p := &importResult.ValidProducts[i]

		// Resolve Category ID
		if p.CategoryID == 0 && p.Category.Name != "" {
			if id, ok := newCategoryIDs[p.Category.Name]; ok {
				p.CategoryID = id
			}
		}

		// Resolve Supplier ID
		if p.SupplierID == 0 && p.Supplier.Name != "" {
			if id, ok := newSupplierIDs[p.Supplier.Name]; ok {
				p.SupplierID = id
			}
		}

		// Resolve Location ID
		if p.LocationID == 0 && p.Location.Name != "" {
			if id, ok := newLocationIDs[p.Location.Name]; ok {
				p.LocationID = id
			}
		}

		// Resolve SubCategory ID
		// Now that we have CategoryID, we can create SubCategory if needed
		if p.SubCategoryID == 0 && p.SubCategory.Name != "" && p.CategoryID != 0 {
			key := fmt.Sprintf("%d::%s", p.CategoryID, strings.ToLower(p.SubCategory.Name))
			if id, ok := createdSubCategories[key]; ok {
				p.SubCategoryID = id
			} else if importResult.NewEntities.SubCategories[p.SubCategory.Name] {
				// It's marked as new, create it
				subCat := &domain.SubCategory{Name: p.SubCategory.Name, CategoryID: p.CategoryID}
				if err := tx.Where("name = ? AND category_id = ?", p.SubCategory.Name, p.CategoryID).FirstOrCreate(subCat).Error; err != nil {
					tx.Rollback()
					c.failJob(job, fmt.Sprintf("failed to create sub-category '%s': %v", p.SubCategory.Name, err))
					return err
				}
				p.SubCategoryID = subCat.ID
				createdSubCategories[key] = subCat.ID
			}
		}

		// Clear nested structs to prevent upsert
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
		d.Nack(false, false)
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
		d.Nack(false, false)
		return
	}

	// Fetch products using filters
	filters := make(map[string]interface{})
	if cat := payload["category"]; cat != nil && cat != "" {
		filters["category"] = cat
	}
	if sup := payload["supplier"]; sup != nil && sup != "" {
		filters["supplier"] = sup
	}

	products, err := c.ProductRepo.GetFiltered(filters)
	if err != nil {
		logrus.Errorf("Failed to get products for export job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := c.saveJob(job); err != nil {
			logrus.Errorf("Failed to update job %d after product fetch error: %v", jobID, err)
		}
		d.Nack(false, false)
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
		d.Nack(false, false)
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
		d.Nack(false, false)
		return
	}

	result := gin.H{
		"downloadUrl": fmt.Sprintf("/api/v1/bulk/files/%s/%s", bucketName, objectName),
	}
	resultBytes, _ := json.Marshal(result)
	job.Result = string(resultBytes)
	job.Status = "COMPLETED"
	if err := c.saveJob(job); err != nil {
		logrus.Errorf("Failed to update job %d: %v", jobID, err)
	}
	d.Ack(false)
}
