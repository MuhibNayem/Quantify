package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rabbitmq/amqp091-go"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"inventory/backend/internal/auth"
	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/notifications"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/router"
	"inventory/backend/internal/services"
	"inventory/backend/internal/websocket"
)

// Payloads for events
type BulkImportJobPayload struct {
	JobID    string `json:"jobId"`
	FilePath string `json:"filePath"`
	UserID   uint   `json:"userId"`
}

type BulkExportJobPayload struct {
	JobID    string `json:"jobId"`
	Format   string `json:"format"`
	Category string `json:"category"`
	Supplier string `json:"supplier"`
	UserID   uint   `json:"userId"`
}

type AlertTriggeredPayload struct {
	ProductID uint   `json:"productId"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

func main() {
	// Initialize Logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	repository.InitDB(cfg)
	defer repository.CloseDB()

	// Initialize Redis client
	repository.InitRedis(cfg)
	defer repository.CloseRedis()

	// Initialize RabbitMQ
	message_broker.InitRabbitMQ(cfg.RabbitMQURL)

	// Initialize JWT
	auth.InitializeJWT(cfg.JWTSecret, cfg.RefreshTokenSecret)

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Register event consumers
	message_broker.Subscribe("inventory", "bulk-import", "bulk.import", hub, func(d amqp091.Delivery) {
		var payload BulkImportJobPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			logrus.Errorf("Failed to unmarshal BulkImportJobPayload: %v", err)
			return
		}

		job := handlers.GetBulkImportJob(payload.JobID)
		if job == nil {
			logrus.Errorf("Job %s not found", payload.JobID)
			return
		}

		if job["status"] == "QUEUED" {
			logrus.Infof("Processing BulkImportJob: %s for file %s by user %d", payload.JobID, payload.FilePath, payload.UserID)

			result, err := services.ProcessBulkImport(payload.FilePath)
			if err != nil {
				job["status"] = "FAILED"
				job["message"] = err.Error()
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to process bulk import: %v", err)
				return
			}

			job["status"] = "PENDING_CONFIRMATION"
			job["totalRecords"] = result.TotalRecords
			job["validRecords"] = result.ValidRecords
			job["invalidRecords"] = result.InvalidRecords
			job["errors"] = result.Errors
			// In a real app, you might store the valid products in a temporary location (e.g., Redis) until confirmation.
			// For simplicity, we'll re-process the file on confirmation.
			handlers.SetBulkImportJob(payload.JobID, job)
			logrus.Infof("BulkImportJob %s validation complete. Status: %s", payload.JobID, job["status"])
		} else if job["status"] == "PROCESSING" {
			logrus.Infof("Confirming BulkImportJob: %s for file %s by user %d", payload.JobID, payload.FilePath, payload.UserID)

			result, err := services.ProcessBulkImport(payload.FilePath)
			if err != nil {
				job["status"] = "FAILED"
				job["message"] = err.Error()
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to process bulk import on confirmation: %v", err)
				return
			}

			if err := repository.DB.CreateInBatches(result.ValidProducts, 100).Error; err != nil {
				job["status"] = "FAILED"
				job["message"] = err.Error()
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to insert products into database: %v", err)
				return
			}

			job["status"] = "COMPLETED"
			job["message"] = "Bulk import completed successfully."
			handlers.SetBulkImportJob(payload.JobID, job)
			logrus.Infof("BulkImportJob %s completed.", payload.JobID)
		}
	})

	message_broker.Subscribe("inventory", "bulk-export", "bulk.export", hub, func(d amqp091.Delivery) {
		var payload BulkExportJobPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			logrus.Errorf("Failed to unmarshal BulkExportJobPayload: %v", err)
			return
		}
		logrus.Infof("Processing BulkExportJob: %s for format %s by user %d", payload.JobID, payload.Format, payload.UserID)

		var products []domain.Product
		db := repository.DB.Preload("Category").Preload("SubCategory").Preload("Supplier")

		if payload.Category != "" {
			db = db.Where("category_id = ?", payload.Category)
		}
		if payload.Supplier != "" {
			db = db.Where("supplier_id = ?", payload.Supplier)
		}

		if err := db.Find(&products).Error; err != nil {
			logrus.Errorf("Failed to fetch products for export: %v", err)
			return
		}

		buffer, err := services.GenerateProductExport(products, payload.Format)
		if err != nil {
			logrus.Errorf("Failed to generate product export: %v", err)
			return
		}

		// In a real app, you would save this file to a cloud storage (e.g., S3)
		// and notify the user with a download link.
		// For now, we'll just log that it's done.
		fileName := fmt.Sprintf("export-%s.%s", payload.JobID, payload.Format)
		os.WriteFile(fileName, buffer.Bytes(), 0644)

		logrus.Infof("BulkExportJob %s completed. Format: %s. File saved to %s", payload.JobID, payload.Format, fileName)
	})

	message_broker.Subscribe("inventory", "alerts", "alert.triggered", hub, func(d amqp091.Delivery) {
		var payload AlertTriggeredPayload
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			logrus.Errorf("Failed to unmarshal AlertTriggeredPayload: %v", err)
			return
		}
		logrus.Infof("Processing AlertTriggeredEvent: %s for product %v. Message: %s", payload.Type, payload.ProductID, payload.Message)

		// In a real scenario, you would have a way to determine which users to notify.
		// For now, we'll assume we notify user with ID 1.
		var userNotificationSettings domain.UserNotificationSettings
		if err := repository.DB.Where("user_id = ?", 1).First(&userNotificationSettings).Error; err != nil {
			logrus.Errorf("Failed to fetch notification settings for user 1: %v", err)
			return
		}

		if userNotificationSettings.EmailNotificationsEnabled {
			subject := fmt.Sprintf("Inventory Alert: %s", payload.Type)
			body := fmt.Sprintf("An alert has been triggered for product ID %d: %s", payload.ProductID, payload.Message)
			if err := notifications.SendEmail(cfg, userNotificationSettings.EmailAddress, subject, body); err != nil {
				logrus.Errorf("Failed to send email notification: %v", err)
			}
		}
	})

	// Initialize Cron Scheduler
	c := cron.New()
	c.AddFunc("@every 5m", func() {
		logrus.Info("Running alert check...")
		handlers.CheckAndTriggerAlerts()
	})
	c.AddFunc("@daily", func() {
		logrus.Info("Running daily sales summary generation...")
		handlers.GenerateDailySalesSummary()
	})
	go c.Start()
	defer c.Stop()

	// Setup router
	r := router.SetupRouter(hub)
	// Start HTTP server
	logrus.Infof("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, r); err != nil {
		logrus.Fatalf("Server failed to start: %v", err)
	}
}
