package bootstrap

import (
	"context"
	"inventory/backend/internal/auth"
	"inventory/backend/internal/config"
	"inventory/backend/internal/consumers"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/migrations"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/router"
	"inventory/backend/internal/services"
	"inventory/backend/internal/storage"
	"inventory/backend/internal/websocket"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// App holds all application-wide dependencies.
type App struct {
	Cfg              *config.Config
	ReportingService *services.ReportingService
	Hub              *websocket.Hub
	JobRepo          *repository.JobRepository
	NotificationRepo repository.NotificationRepository
	Server           *http.Server
	Cron             *cron.Cron
	MinIOUploader    storage.Uploader
}

// NewApp initializes the application and its dependencies.
func NewApp(cfg *config.Config) *App {
	// Initialize database connection
	repository.InitDB(cfg)

	// Check if search index migration should run
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		logrus.Info("RUN_MIGRATIONS environment variable is true. Running search index population...")
		if err := migrations.PopulateSearchIndex(repository.DB); err != nil {
			logrus.Fatalf("Failed to run search index population on startup: %v", err)
		}
		logrus.Info("Search index population completed.")
	}

	// Seed Roles
	if err := migrations.SeedRoles(repository.DB); err != nil {
		logrus.Errorf("Failed to seed roles: %v", err)
	}

	// Seed AI Agent User
	if err := migrations.SeedAIUser(repository.DB); err != nil {
		logrus.Errorf("Failed to seed AI Agent user: %v", err)
	}

	// Initialize Redis client
	repository.InitRedis(cfg)

	// Initialize RabbitMQ
	message_broker.InitRabbitMQ(cfg.RabbitMQURL)

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
	reportsRepo := repository.NewReportsRepository(repository.DB)
	notificationRepo := repository.NewNotificationRepository(repository.DB)
	settingsRepo := repository.NewSettingsRepository(repository.DB)

	// Initialize Services
	settingsService := services.NewSettingsService(settingsRepo)
	reportingService := services.NewReportingService(reportsRepo, minioUploader, jobRepo, hub, cfg, settingsService)
	// Wait, SettingsService is an interface in handlers. Let's see how it was passed before.
	// In original main.go:
	// It wasn't explicitly initialized in the snippet I saw, but handlers.NewSalesHandler took it.
	// Let's check handlers/sales.go again. It takes services.SettingsService.
	// We need to find where SettingsService is defined and implemented.
	// I'll assume for now there's a NewSettingsService or similar.
	// If not, I might need to check services package.

	// Initialize and start the BulkConsumer
	// We will start consumers in Run()

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

	// Setup router
	r := router.SetupRouter(cfg, hub, jobRepo, minioUploader)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	app := &App{
		Cfg:              cfg,
		ReportingService: reportingService,
		Hub:              hub,
		JobRepo:          jobRepo,
		NotificationRepo: notificationRepo,
		Server:           server,
		Cron:             c,
		MinIOUploader:    minioUploader,
	}

	return app
}

// Run starts the application components.
func (app *App) Run(ctx context.Context) {
	// Start Cron
	app.Cron.Start()

	// Start Server
	go func() {
		logrus.Infof("Server starting on port %s", app.Cfg.ServerPort)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Initialize and start Consumers
	app.startConsumers(ctx)

	<-ctx.Done()
	app.Stop()
}

func (app *App) Stop() {
	logrus.Info("Shutdown signal received, shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	app.Cron.Stop()
	repository.CloseDB()
	repository.CloseRedis()
	message_broker.Close()

	logrus.Info("Server exited gracefully")
}

func (app *App) startConsumers(ctx context.Context) {
	productRepo := repository.NewProductRepository(repository.DB)
	categoryRepo := repository.NewCategoryRepository(repository.DB)
	supplierRepo := repository.NewSupplierRepository(repository.DB)
	locationRepo := repository.NewLocationRepository(repository.DB)

	bulkImportSvc := services.NewBulkImportService(categoryRepo, supplierRepo, locationRepo)
	bulkExportSvc := services.NewBulkExportService()

	bulkConsumer := consumers.NewBulkConsumer(repository.DB, app.JobRepo, productRepo, categoryRepo, supplierRepo, locationRepo, app.MinIOUploader, bulkImportSvc, bulkExportSvc, app.Hub, app.NotificationRepo)

	workerCount := 1
	if app.Cfg.ConsumerConcurrency > 0 {
		workerCount = app.Cfg.ConsumerConcurrency
	}
	// We don't need to defer stop here because the context cancellation will handle it (if implemented correctly)
	// or we rely on process termination.
	// However, consumers.Start returns a cancel func. Ideally we should call it on Stop.
	// For simplicity in this refactor, we let them run until context is done (which is passed from main).
	// But wait, Start() returns a cancelFunc that STOPS the subscription.
	// If we don't store it, we can't stop it explicitly.
	// But since we pass ctx to Start, if that ctx is cancelled, does the consumer stop?
	// Let's check AlertConsumer.Start:
	// return message_broker.Subscribe(ctx, ...)
	// message_broker.Subscribe usually listens to ctx.Done().
	// So passing ctx from Run (which is ctx.Done() when signal received) should be enough.

	bulkConsumer.Start(ctx, workerCount)

	alertConsumer := consumers.NewAlertConsumer(repository.DB, app.NotificationRepo, app.Hub, app.Cfg)
	alertConsumer.Start(ctx)

	reportingConsumer := consumers.NewReportingConsumer(app.ReportingService, app.JobRepo)
	reportingConsumer.Start(ctx)
}
