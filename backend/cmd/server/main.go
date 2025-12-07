package main

import (
	"context"
	"encoding/json"
	"inventory/backend/internal/consumers"
	"inventory/backend/internal/migrations"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"inventory/backend/internal/storage"
	"inventory/backend/internal/websocket"
)

// App holds all application-wide dependencies.
type App struct {
	cfg              *config.Config
	reportingService *services.ReportingService
	hub              *websocket.Hub
	jobRepo          *repository.JobRepository
	notificationRepo repository.NotificationRepository
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	repository.InitDB(cfg)
	defer repository.CloseDB()

	// Check if search index migration should run
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		logrus.Info("RUN_MIGRATIONS environment variable is true. Running search index population...")
		if err := migrations.PopulateSearchIndex(repository.DB); err != nil {
			logrus.Fatalf("Failed to run search index population on startup: %v", err)
		}
		logrus.Info("Search index population completed.")
	}

	// Initialize Redis client
	repository.InitRedis(cfg)
	defer repository.CloseRedis()

	// Initialize RabbitMQ
	message_broker.InitRabbitMQ(cfg.RabbitMQURL)
	defer message_broker.Close()

	// Initialize JWT
	auth.InitializeJWT(cfg.JWTSecret, cfg.RefreshTokenSecret)

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize MinIO Uploader
	minioUploader, err := storage.NewMinIOUploader(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize MinIO uploader: %v", err)
	}

	// Initialize Repositories
	jobRepo := repository.NewJobRepository(repository.DB)
	productRepo := repository.NewProductRepository(repository.DB)
	reportsRepo := repository.NewReportsRepository(repository.DB)
	notificationRepo := repository.NewNotificationRepository(repository.DB)
	categoryRepo := repository.NewCategoryRepository(repository.DB)
	supplierRepo := repository.NewSupplierRepository(repository.DB)
	locationRepo := repository.NewLocationRepository(repository.DB)

	// Initialize Services
	bulkImportSvc := services.NewBulkImportService(categoryRepo, supplierRepo, locationRepo)
	bulkExportSvc := services.NewBulkExportService()
	reportingService := services.NewReportingService(reportsRepo, minioUploader, jobRepo, cfg)

	app := &App{
		cfg:              cfg,
		reportingService: reportingService,
		hub:              hub,
		jobRepo:          jobRepo,
		notificationRepo: notificationRepo,
	}

	// Initialize and start the BulkConsumer
	bulkConsumer := consumers.NewBulkConsumer(repository.DB, jobRepo, productRepo, categoryRepo, supplierRepo, locationRepo, minioUploader, bulkImportSvc, bulkExportSvc, hub)
	workerCount := 1
	if cfg.ConsumerConcurrency > 0 {
		workerCount = cfg.ConsumerConcurrency
	}
	stopConsumers := bulkConsumer.Start(ctx, workerCount)
	defer stopConsumers()

	alertConsumer := consumers.NewAlertConsumer()
	stopAlertConsumer := alertConsumer.Start(ctx)
	defer stopAlertConsumer()

	message_broker.Subscribe(ctx, "inventory", "alerts", "alert.triggered", func(ctx context.Context, deliveries <-chan amqp091.Delivery) {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				app.handleAlertDelivery(d)
			}
		}
	})

	message_broker.Subscribe(ctx, "inventory", "reporting", "report.generate", func(ctx context.Context, deliveries <-chan amqp091.Delivery) {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				app.handleReportingDelivery(d)
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
		reportingService.GenerateDailySalesSummary()
	})
	go c.Start()
	defer c.Stop()

	// Setup router
	r := router.SetupRouter(cfg, hub, jobRepo, minioUploader)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		logrus.Infof("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-ctx.Done()
	logrus.Info("Shutdown signal received, shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited gracefully")
}

func (app *App) handleAlertDelivery(d amqp091.Delivery) {
	var payload handlers.AlertTriggeredPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		logrus.Errorf("Failed to unmarshal alert payload: %v", err)
		d.Nack(false, false)
		return
	}

	logrus.Infof("Handling alert delivery for alert type %s", payload.Type)

	// 1. Find roles subscribed to this alert type
	var subscriptions []domain.AlertRoleSubscription
	if err := repository.DB.Where("alert_type = ?", payload.Type).Find(&subscriptions).Error; err != nil {
		logrus.Errorf("Failed to fetch alert subscriptions: %v", err)
		d.Nack(true, true) // Nack and requeue
		return
	}

	if len(subscriptions) == 0 {
		logrus.Infof("No subscriptions found for alert type %s", payload.Type)
		d.Ack(false) // No one is subscribed, so we're done.
		return
	}

	var roles []string
	for _, sub := range subscriptions {
		roles = append(roles, sub.Role)
	}

	// 2. Find users with those roles
	var users []domain.User
	if err := repository.DB.Where("role IN (?)", roles).Find(&users).Error; err != nil {
		logrus.Errorf("Failed to fetch users for roles %v: %v", roles, err)
		d.Nack(true, true)
		return
	}

	// 3. Send notifications to the targeted users
	for _, user := range users {
		var settings domain.UserNotificationSettings
		if err := repository.DB.Where("user_id = ?", user.ID).First(&settings).Error; err != nil {
			// If settings don't exist, we can't send notifications, so we skip.
			continue
		}

		// Send email if enabled
		if settings.EmailNotificationsEnabled && settings.EmailAddress != "" {
			subject := "Inventory Alert: " + payload.Type
			body := payload.Message
			if err := notifications.SendEmail(app.cfg, settings.EmailAddress, subject, body); err != nil {
				logrus.Errorf("Failed to send email to %s: %v", settings.EmailAddress, err)
			}
		}

		// Create and persist in-app notification
		notificationPayload, _ := json.Marshal(payload)
		notification := domain.Notification{
			UserID:      user.ID,
			Type:        "ALERT",
			Title:       "New Alert: " + payload.Type,
			Message:     payload.Message,
			Payload:     string(notificationPayload),
			TriggeredAt: time.Now(),
		}
		if err := app.notificationRepo.CreateNotification(&notification); err != nil {
			logrus.Errorf("Failed to create in-app notification for user %d: %v", user.ID, err)
		} else {
			// Send user-specific WebSocket message
			app.hub.SendToUser(user.ID, notification)
		}
	}

	d.Ack(false)
}

func (app *App) handleReportingDelivery(d amqp091.Delivery) {
	var job domain.Job
	if err := json.Unmarshal(d.Body, &job); err != nil {
		logrus.Errorf("Failed to unmarshal job payload: %v", err)
		d.Nack(false, false) // Negative acknowledgement, don't requeue
		return
	}

	logrus.Infof("Handling reporting delivery for job %d: %s", job.ID, job.Type)

	if err := app.reportingService.GenerateReport(&job); err != nil {
		logrus.Errorf("Failed to generate report for job %d: %v", job.ID, err)
		job.Status = "FAILED"
		job.LastError = err.Error()
		if err := app.jobRepo.UpdateJob(&job); err != nil {
			logrus.Errorf("Failed to update job status for job %d: %v", job.ID, err)
		}
		d.Nack(false, false) // Nack and don't requeue
		return
	}

	d.Ack(false) // Acknowledge the message
}
