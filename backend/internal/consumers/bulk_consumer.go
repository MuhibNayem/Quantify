package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
	"inventory/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type BulkConsumer struct {
	JobRepo       *repository.JobRepository
	ProductRepo   *repository.ProductRepository
	Uploader      storage.Uploader
	BulkImportSvc *services.BulkImportService
	BulkExportSvc *services.BulkExportService
}

func NewBulkConsumer(jobRepo *repository.JobRepository, productRepo *repository.ProductRepository, uploader storage.Uploader, bulkImportSvc *services.BulkImportService, bulkExportSvc *services.BulkExportService) *BulkConsumer {
	return &BulkConsumer{
		JobRepo:       jobRepo,
		ProductRepo:   productRepo,
		Uploader:      uploader,
		BulkImportSvc: bulkImportSvc,
		BulkExportSvc: bulkExportSvc,
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
	c.JobRepo.UpdateJob(job)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		logrus.Errorf("Failed to unmarshal job payload: %v", err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	bucketName := payload["bucketName"].(string)
	objectName := payload["objectName"].(string)

	file, err := c.Uploader.DownloadFile(bucketName, objectName)
	if err != nil {
		logrus.Errorf("Failed to download file %s from bucket %s: %v", objectName, bucketName, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}
	defer file.Close()

	result, err := c.BulkImportSvc.ProcessBulkImport(file)
	if err != nil {
		logrus.Errorf("Failed to process bulk import for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	resultBytes, _ := json.Marshal(result)
	job.Result = string(resultBytes)
	job.Status = "PENDING_CONFIRMATION"
	if err := c.JobRepo.UpdateJob(job); err != nil {
		logrus.Errorf("Failed to update job %d: %v", jobID, err)
	}
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

	if job.Status != "PROCESSING" {
		logrus.Warnf("Job %d is not in PROCESSING state, skipping confirmation", jobID)
		return
	}

	var importResult services.BulkImportResult
	if err := json.Unmarshal([]byte(job.Result), &importResult); err != nil {
		logrus.Errorf("Failed to unmarshal import result for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	if err := c.ProductRepo.CreateBulk(&importResult.ValidProducts); err != nil {
		logrus.Errorf("Failed to bulk insert products for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	job.Status = "COMPLETED"
	if err := c.JobRepo.UpdateJob(job); err != nil {
		logrus.Errorf("Failed to update job %d: %v", jobID, err)
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
	c.JobRepo.UpdateJob(job)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		logrus.Errorf("Failed to unmarshal job payload: %v", err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	// Fetch products from the database
	// This is a simplified example. In a real application, you would use filters from the payload.
	products, err := c.ProductRepo.GetAll()
	if err != nil {
		logrus.Errorf("Failed to get products for export job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	format := payload["format"].(string)
	buffer, err := c.BulkExportSvc.GenerateProductExport(products, format)
	if err != nil {
		logrus.Errorf("Failed to generate export file for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	bucketName := "bulk-exports"
	objectName := fmt.Sprintf("%d-export.%s", jobID, format)
	_, err = c.Uploader.UploadFile(bucketName, objectName, buffer, int64(buffer.Len()))
	if err != nil {
		logrus.Errorf("Failed to upload export file for job %d: %v", jobID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		c.JobRepo.UpdateJob(job)
		return
	}

	result := gin.H{
		"downloadUrl": fmt.Sprintf("/files/%s/%s", bucketName, objectName),
	}
	resultBytes, _ := json.Marshal(result)
	job.Result = string(resultBytes)
	job.Status = "COMPLETED"
	if err := c.JobRepo.UpdateJob(job); err != nil {
		logrus.Errorf("Failed to update job %d: %v", jobID, err)
	}
}
