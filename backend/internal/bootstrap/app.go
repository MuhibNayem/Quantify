package bootstrap

import (
	"context"
	"inventory/backend/internal/app"
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
	Container        *app.Container
	Cfg              *config.Config
	ReportingService *services.ReportingService
	Hub              *websocket.Hub
	JobRepo          *repository.JobRepository
	NotificationRepo repository.NotificationRepository
	Server           *http.Server
	Cron             *cron.Cron
	MinIOUploader    storage.Uploader

	// cancelFuncs holds the cancel funcs returned by each consumer's
	// Start(). Stop() invokes them so the AMQP Subscribe goroutines exit
	// before we tear down the message-broker connection.
	cancelFuncs []context.CancelFunc
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
	c := &app.Container{
		Cfg:      cfg,
		DB:       repository.DB,
		Redis:    repository.GetClient(),
		Hub:      hub,
		Uploader: minioUploader,
		Jobs:     jobRepo,
	}
	c.Products = repository.NewProductRepository(c.DB)
	c.Categories = repository.NewCategoryRepository(c.DB)
	c.Suppliers = repository.NewSupplierRepository(c.DB)
	c.Payments = repository.NewPaymentRepository(c.DB)
	c.Forecasting = repository.NewForecastingRepository(c.DB)
	c.Barcode = repository.NewBarcodeRepository(c.DB)
	c.CRM = repository.NewCRMRepository(c.DB)
	c.TimeTracking = repository.NewTimeTrackingRepository(c.DB)
	c.Reports = repository.NewReportsRepository(c.DB)
	c.Notifications = repository.NewNotificationRepository(c.DB)
	c.Search = repository.NewSearchRepository(c.DB)
	c.Users = repository.NewUserRepository(c.DB)
	c.Dashboard = repository.NewDashboardRepository(c.DB)
	c.Replenishment = repository.NewReplenishmentRepository(c.DB)
	c.SettingsRepo = repository.NewSettingsRepository(c.DB)
	c.Roles = repository.NewRoleRepository(c.DB, c.Redis)
	c.Locations = repository.NewLocationRepository(c.DB)

	// Initialize Services
	c.Payment = services.NewPaymentService(cfg, c.Payments)
	c.ForecastingSvc = services.NewForecastingService(c.Forecasting, cfg)
	c.BarcodeSvc = services.NewBarcodeService(c.Barcode)
	c.SettingsSvc = services.NewSettingsService(c.SettingsRepo)
	c.CRMService = services.NewCRMService(c.CRM, c.DB, c.SettingsSvc, cfg)
	c.TimeTrackingSvc = services.NewTimeTrackingService(c.TimeTracking, c.Users)
	c.Integration = services.NewIntegrationService()
	c.Reporting = services.NewReportingService(c.Reports, c.Uploader, c.Jobs, c.Hub, cfg, c.SettingsSvc)
	c.SearchSvc = services.NewSearchService(c.DB, c.Search, c.Products, c.Users, c.Suppliers, c.Categories)
	c.ReplenishmentSvc = services.NewReplenishmentService(c.Replenishment)
	c.Role = services.NewRoleService(c.Roles)
	c.Sales = services.NewSalesService(c.DB, c.SettingsSvc, c.Reporting)
	c.Email = services.NewEmailService(cfg)
	c.PurchaseOrder = services.NewPurchaseOrderService(c.DB, c.Email)
	c.VendorReturn = services.NewVendorReturnService(c.DB)

	// Initialize Handlers
	c.Product = handlers.NewProductHandler(c.Products, c.DB)
	c.Category = handlers.NewCategoryHandler(c.Categories, c.DB)
	c.Supplier = handlers.NewSupplierHandler(c.Suppliers, c.DB)
	c.PaymentHandler = handlers.NewPaymentHandler(c.Payment)
	c.ReplenishmentHandler = handlers.NewReplenishmentHandler(c.ForecastingSvc, c.ReplenishmentSvc, c.PurchaseOrder, c.VendorReturn, c.Hub, c.Notifications, c.Email)
	c.BarcodeHandler = handlers.NewBarcodeHandler(c.BarcodeSvc)
	c.CRMHandler = handlers.NewCRMHandler(c.CRMService, c.DB)
	c.TimeTrackingHandler = handlers.NewTimeTrackingHandler(c.TimeTrackingSvc)
	c.Webhook = handlers.NewWebhookHandler(c.Integration)
	c.Report = handlers.NewReportHandler(c.Reporting, c.Jobs)
	c.Bulk = handlers.NewBulkHandler(c.Jobs, c.Uploader)
	c.Notification = handlers.NewNotificationHandler(c.Notifications).WithDB(c.DB)
	c.User = handlers.NewUserHandler(c.Users, cfg)
	c.SearchHandler = handlers.NewSearchHandler(c.SearchSvc)
	c.DashboardHandler = handlers.NewDashboardHandler(c.Dashboard)
	c.SettingsHandler = handlers.NewSettingsHandler(c.SettingsSvc)
	c.RoleHandler = handlers.NewRoleHandler(c.Role)
	c.Return = handlers.NewReturnHandler(c.DB, cfg, c.SettingsSvc, c.Hub, c.Notifications, c.Reporting)
	c.Promotion = handlers.NewPromotionHandler(c.DB)
	c.SalesHandler = handlers.NewSalesHandler(c.DB, c.SettingsSvc, c.Reporting, c.Sales)

	reportingService := c.Reporting
	notificationRepo := c.Notifications

	// Initialize and start the BulkConsumer
	// We will start consumers in Run()

	// Initialize Cron Scheduler
	scheduler := cron.New()
	scheduler.AddFunc("@every 5m", func() {
		logrus.Info("Running alert check...")
		handlers.CheckAndTriggerAlerts()
	})

	scheduler.AddFunc("@daily", func() {
		logrus.Info("Running daily sales summary generation...")
		reportingService.GenerateDailySalesSummary()
	})

	// Setup router
	r := router.SetupRouter(c)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	app := &App{
		Container:        c,
		Cfg:              cfg,
		ReportingService: reportingService,
		Hub:              hub,
		JobRepo:          jobRepo,
		NotificationRepo: notificationRepo,
		Server:           server,
		Cron:             scheduler,
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

	// Cancel consumer subscriptions first so the AMQP delivery loops
	// exit before we close the broker connection underneath them.
	for _, cancel := range app.cancelFuncs {
		cancel()
	}
	app.cancelFuncs = nil

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

	if bulkCancel := bulkConsumer.Start(ctx, workerCount); bulkCancel != nil {
		app.cancelFuncs = append(app.cancelFuncs, bulkCancel)
	}

	alertConsumer := consumers.NewAlertConsumer(repository.DB, app.NotificationRepo, app.Hub, app.Cfg)
	alertCancel := alertConsumer.Start(ctx)
	if alertCancel != nil {
		app.cancelFuncs = append(app.cancelFuncs, alertCancel)
	}

	reportingConsumer := consumers.NewReportingConsumer(app.ReportingService, app.JobRepo)
	reportingCancel := reportingConsumer.Start(ctx)
	if reportingCancel != nil {
		app.cancelFuncs = append(app.cancelFuncs, reportingCancel)
	}
}
