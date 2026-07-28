package app

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"inventory/backend/internal/config"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
	"inventory/backend/internal/storage"
	"inventory/backend/internal/websocket"
)

// Container holds all wired dependencies for the application. Built once
// in bootstrap, consumed by router/handlers/services. Replaces ad-hoc
// package-level globals (repository.DB) so tests can inject fakes.
type Container struct {
	Cfg      *config.Config
	DB       *gorm.DB
	Redis    redis.UniversalClient
	AMQP     *amqp.Connection
	Hub      *websocket.Hub
	Uploader storage.Uploader

	Roles         repository.RoleRepository
	Users         *repository.UserRepository
	Products      *repository.ProductRepository
	Categories    *repository.CategoryRepository
	Suppliers     *repository.SupplierRepository
	Locations     *repository.LocationRepository
	Jobs          *repository.JobRepository
	Reports       *repository.ReportsRepository
	Notifications repository.NotificationRepository
	SettingsRepo  repository.SettingsRepository
	Payments      repository.PaymentRepository
	TimeTracking  repository.TimeTrackingRepository
	Barcode       repository.BarcodeRepository
	Search        *repository.SearchRepository
	CRM           repository.CRMRepository
	Replenishment repository.ReplenishmentRepository
	Forecasting   repository.ForecastingRepository
	Dashboard     *repository.DashboardRepository

	Role             services.RoleService
	SettingsSvc      services.SettingsService
	Reporting        *services.ReportingService
	Sales            *services.SalesService
	StockTransfer    *services.StockTransferService
	SearchSvc        *services.SearchService
	BulkImport       *services.BulkImportService
	BulkExport       *services.BulkExportService
	VendorReturn     services.VendorReturnService
	ForecastingSvc   services.ForecastingService
	Integration      *services.IntegrationService
	ReplenishmentSvc services.ReplenishmentService
	CRMService       services.CRMService
	BarcodeSvc       services.BarcodeService
	Email            services.EmailService
	Payment          services.PaymentService
	PurchaseOrder    services.PurchaseOrderService
	TimeTrackingSvc  services.TimeTrackingService

	User                 *handlers.UserHandler
	Product              *handlers.ProductHandler
	Category             *handlers.CategoryHandler
	Supplier             *handlers.SupplierHandler
	RoleHandler          *handlers.RoleHandler
	Report               *handlers.ReportHandler
	Notification         *handlers.NotificationHandler
	SettingsHandler      *handlers.SettingsHandler
	PaymentHandler       *handlers.PaymentHandler
	TimeTrackingHandler  *handlers.TimeTrackingHandler
	BarcodeHandler       *handlers.BarcodeHandler
	SearchHandler        *handlers.SearchHandler
	CRMHandler           *handlers.CRMHandler
	ReplenishmentHandler *handlers.ReplenishmentHandler
	Bulk                 *handlers.BulkHandler
	SalesHandler         *handlers.SalesHandler
	Return               *handlers.ReturnHandler
	Promotion            *handlers.PromotionHandler
	DashboardHandler     *handlers.DashboardHandler
	Webhook              *handlers.WebhookHandler
}
