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

			// Store valid products in Redis with a TTL of 1 hour
			validProductsJSON, err := json.Marshal(result.ValidProducts)
			if err != nil {
				job["status"] = "FAILED"
				job["message"] = "Failed to serialize valid products"
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to serialize valid products: %v", err)
				return
			}
			err = repository.SetCache("bulk_import:"+payload.JobID, string(validProductsJSON), 3600)
			if err != nil {
				job["status"] = "FAILED"
				job["message"] = "Failed to store valid products in cache"
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to store valid products in cache: %v", err)
				return
			}

			job["status"] = "PENDING_CONFIRMATION"
			job["totalRecords"] = result.TotalRecords
			job["validRecords"] = result.ValidRecords
			job["invalidRecords"] = result.InvalidRecords
			job["errors"] = result.Errors
			handlers.SetBulkImportJob(payload.JobID, job)
			logrus.Infof("BulkImportJob %s validation complete. Status: %s", payload.JobID, job["status"])
		} else if job["status"] == "PROCESSING" {
			logrus.Infof("Confirming BulkImportJob: %s for file %s by user %d", payload.JobID, payload.FilePath, payload.UserID)

			// Retrieve valid products from Redis
			validProductsJSON, err := repository.GetCache("bulk_import:" + payload.JobID)
			if err != nil {
				job["status"] = "FAILED"
				job["message"] = "Failed to retrieve valid products from cache"
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to retrieve valid products from cache: %v", err)
				return
			}
			if validProductsJSON == "" {
				job["status"] = "FAILED"
				job["message"] = "No valid products found in cache"
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("No valid products found in cache for job %s", payload.JobID)
				return
			}

			var validProducts []domain.Product
			if err := json.Unmarshal([]byte(validProductsJSON), &validProducts); err != nil {
				job["status"] = "FAILED"
				job["message"] = "Failed to deserialize valid products"
				handlers.SetBulkImportJob(payload.JobID, job)
				logrus.Errorf("Failed to deserialize valid products: %v", err)
				return
			}

			if err := repository.DB.CreateInBatches(validProducts, 100).Error; err != nil {
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

			// Delete the cached data
			err = repository.DeleteCache("bulk_import:" + payload.JobID)
			if err != nil {
				logrus.Errorf("Failed to delete cached bulk import data for job %s: %v", payload.JobID, err)
			}
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

		// Notify all admin users
		var adminUsers []domain.User
		if err := repository.DB.Where("role = ?", "Admin").Find(&adminUsers).Error; err != nil {
			logrus.Errorf("Failed to fetch admin users: %v", err)
			return
		}

		for _, user := range adminUsers {
			var userNotificationSettings domain.UserNotificationSettings
			if err := repository.DB.Where("user_id = ?", user.ID).First(&userNotificationSettings).Error; err != nil {
				logrus.Errorf("Failed to fetch notification settings for user %d: %v", user.ID, err)
				continue
			}

			if userNotificationSettings.EmailNotificationsEnabled {
				subject := fmt.Sprintf("Inventory Alert: %s", payload.Type)
				body := fmt.Sprintf("An alert has been triggered for product ID %d: %s", payload.ProductID, payload.Message)
				if err := notifications.SendEmail(cfg, userNotificationSettings.EmailAddress, subject, body); err != nil {
					logrus.Errorf("Failed to send email notification to user %d: %v", user.ID, err)
				}
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
