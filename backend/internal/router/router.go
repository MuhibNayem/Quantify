package router

import (
	"inventory/backend/internal/config"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/services"
	"inventory/backend/internal/storage"
	"inventory/backend/internal/websocket"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "inventory/backend/api"
)

func SetupRouter(cfg *config.Config, hub *websocket.Hub, jobRepo *repository.JobRepository, minioUploader storage.Uploader) *gin.Engine {
	r := gin.Default()

	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Authorization", "X-Requested-With", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowWebSockets:  true,
	}

	r.Use(cors.New(corsConfig))

	r.Use(middleware.ErrorHandler())

	// Initialize repositories
	db := repository.DB
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	forecastingRepo := repository.NewForecastingRepository(db)
	barcodeRepo := repository.NewBarcodeRepository(db)
	crmRepo := repository.NewCRMRepository(db)
	timeTrackingRepo := repository.NewTimeTrackingRepository(db)
	reportsRepo := repository.NewReportsRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	searchRepo := repository.NewSearchRepository(db)
	userRepo := repository.NewUserRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	replenishmentRepo := repository.NewReplenishmentRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	roleRepo := repository.NewRoleRepository(db, repository.GetClient())

	// Initialize services
	paymentService := services.NewPaymentService(cfg, paymentRepo)
	forecastingService := services.NewForecastingService(forecastingRepo)
	barcodeService := services.NewBarcodeService(barcodeRepo)
	settingsService := services.NewSettingsService(settingsRepo)
	crmService := services.NewCRMService(crmRepo, db, settingsService)
	timeTrackingService := services.NewTimeTrackingService(timeTrackingRepo, userRepo)
	integrationService := services.NewIntegrationService()
	reportingService := services.NewReportingService(reportsRepo, minioUploader, jobRepo, cfg)
	searchService := services.NewSearchService(db, searchRepo, productRepo, userRepo, supplierRepo, categoryRepo)
	replenishmentService := services.NewReplenishmentService(replenishmentRepo)
	roleService := services.NewRoleService(roleRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productRepo, db)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo, db)
	supplierHandler := handlers.NewSupplierHandler(supplierRepo, db)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	replenishmentHandler := handlers.NewReplenishmentHandler(forecastingService, replenishmentService)
	barcodeHandler := handlers.NewBarcodeHandler(barcodeService)
	crmHandler := handlers.NewCRMHandler(crmService)
	timeTrackingHandler := handlers.NewTimeTrackingHandler(timeTrackingService)
	webhookHandler := handlers.NewWebhookHandler(integrationService)
	reportHandler := handlers.NewReportHandler(reportingService, jobRepo)
	bulkHandler := handlers.NewBulkHandler(jobRepo, minioUploader)
	notificationHandler := handlers.NewNotificationHandler(notificationRepo)
	userHandler := handlers.NewUserHandler(userRepo, db)
	searchHandler := handlers.NewSearchHandler(searchService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardRepo)
	settingsHandler := handlers.NewSettingsHandler(settingsService)
	roleHandler := handlers.NewRoleHandler(roleService)
	returnHandler := handlers.NewReturnHandler(db, cfg, settingsService)

	// Public routes (no tenant middleware)
	publicRoutes := r.Group("/")
	{
		publicRoutes.GET("/health", handlers.HealthCheck)
		publicRoutes.GET("/ws", func(c *gin.Context) {
			handlers.ServeWs(hub, c, notificationRepo)
		})
		publicRoutes.GET("/metrics", gin.WrapH(promhttp.Handler()))
		// Swagger UI
		publicRoutes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Webhook routes
	webhookRoutes := r.Group("/webhooks")
	{
		webhookRoutes.POST("", webhookHandler.HandleWebhook)
	}

	// Payment routes
	paymentRoutes := r.Group("/payment")
	{
		paymentRoutes.POST("/create", paymentHandler.CreatePayment)
		paymentRoutes.POST("/success", paymentHandler.PaymentSuccess)
		paymentRoutes.POST("/fail", paymentHandler.PaymentFail)
		paymentRoutes.POST("/cancel", paymentHandler.PaymentCancel)
		paymentRoutes.POST("/ipn", paymentHandler.HandleSSLCommerzIPN) // Updated IPN handler
	}

	// Public API routes (no auth middleware)
	publicAPI := r.Group("/api/v1")
	{
		userRoutes := publicAPI.Group("/users")
		{
			userRoutes.POST("/register", userHandler.RegisterUser)
			userRoutes.POST("/login", userHandler.LoginUser)
		}
	}

	api := r.Group("/api/v1")
	{
		// Auth middleware will be applied to all routes in this group
		api.Use(middleware.AuthMiddleware())
		api.Use(middleware.CSRFMiddleware())

		// Jobs (Manager/Admin)
		jobs := api.Group("/jobs")
		jobs.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
		{
			jobs.POST("/:jobId/cancel", reportHandler.CancelJob)
		}

		// Dashboard
		api.GET("/dashboard/summary", middleware.RequirePermission(roleRepo, "dashboard.view"), dashboardHandler.GetDashboardSummary)

		// Products
		products := api.Group("/products")
		{
			products.POST("", middleware.RequirePermission(roleRepo, "products.write"), productHandler.CreateProduct)
			products.GET("", middleware.RequirePermission(roleRepo, "products.read"), productHandler.ListProducts)
			products.GET("/:productId", middleware.RequirePermission(roleRepo, "products.read"), productHandler.GetProduct)
			products.GET("/sku/:sku", middleware.RequirePermission(roleRepo, "products.read"), productHandler.GetProductBySKU)
			products.GET("/barcode/:barcode", middleware.RequirePermission(roleRepo, "products.read"), productHandler.GetProductByBarcode)
			products.PUT("/:productId", middleware.RequirePermission(roleRepo, "products.write"), productHandler.UpdateProduct)
			products.DELETE("/:productId", middleware.RequirePermission(roleRepo, "products.delete"), productHandler.DeleteProduct)
			products.GET("/:productId/stock", middleware.RequirePermission(roleRepo, "products.read"), handlers.GetProductStock)
			products.POST("/:productId/stock/batches", middleware.RequirePermission(roleRepo, "products.write"), handlers.CreateBatch)
			products.POST("/:productId/stock/adjustments", middleware.RequirePermission(roleRepo, "products.write"), handlers.CreateStockAdjustment)
			products.GET("/:productId/history", middleware.RequirePermission(roleRepo, "products.read"), handlers.ListStockHistory)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.POST("", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.CreateCategory)
			categories.GET("", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.ListCategories)
			categories.GET("/:categoryId", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.GetCategory)
			categories.GET("/name/:name", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.GetCategoryByName)
			categories.PUT("/:categoryId", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.UpdateCategory)
			categories.DELETE("/:categoryId", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.DeleteCategory)
			categories.POST("/:categoryId/sub-categories", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.CreateSubCategory)
			categories.GET("/:categoryId/sub-categories", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.ListSubCategories)
		}

		// Sub-categories
		subCategories := api.Group("/sub-categories")
		{
			subCategories.PUT("/:id", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.UpdateSubCategory)
			subCategories.DELETE("/:id", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.DeleteSubCategory)
		}

		// Suppliers
		suppliers := api.Group("/suppliers")
		{
			suppliers.POST("", middleware.RequirePermission(roleRepo, "suppliers.write"), supplierHandler.CreateSupplier)
			suppliers.GET("", middleware.RequirePermission(roleRepo, "suppliers.read"), supplierHandler.ListSuppliers)
			suppliers.GET("/:id", middleware.RequirePermission(roleRepo, "suppliers.read"), supplierHandler.GetSupplier)
			suppliers.GET("/name/:name", middleware.RequirePermission(roleRepo, "suppliers.read"), supplierHandler.GetSupplierByName)
			suppliers.PUT("/:id", middleware.RequirePermission(roleRepo, "suppliers.write"), supplierHandler.UpdateSupplier)
			suppliers.DELETE("/:id", middleware.RequirePermission(roleRepo, "suppliers.write"), supplierHandler.DeleteSupplier)
			suppliers.GET("/:id/performance", middleware.RequirePermission(roleRepo, "reports.financial"), supplierHandler.GetSupplierPerformanceReport)
		}

		// Locations
		locations := api.Group("/locations")
		{
			locations.POST("", middleware.RequirePermission(roleRepo, "locations.write"), handlers.CreateLocation)
			locations.GET("", middleware.RequirePermission(roleRepo, "locations.read"), handlers.ListLocations)
			locations.GET("/:id", middleware.RequirePermission(roleRepo, "locations.read"), handlers.GetLocation)
			locations.PUT("/:id", middleware.RequirePermission(roleRepo, "locations.write"), handlers.UpdateLocation)
			locations.DELETE("/:id", middleware.RequirePermission(roleRepo, "locations.write"), handlers.DeleteLocation)
		}

		// Barcode
		barcode := api.Group("/barcode")
		{
			barcode.GET("/lookup", middleware.RequirePermission(roleRepo, "barcode.read"), barcodeHandler.LookupProductByBarcode)
			barcode.GET("/generate", middleware.RequirePermission(roleRepo, "barcode.read"), barcodeHandler.GenerateBarcode)
		}

		// Replenishment
		replenishment := api.Group("/replenishment")
		{
			replenishment.POST("/forecast/generate", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.GenerateDemandForecast)
			replenishment.POST("/suggestions/generate", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.GenerateReorderSuggestions)
			replenishment.GET("/forecast/:forecastId", middleware.RequirePermission(roleRepo, "replenishment.read"), handlers.GetDemandForecast)
			replenishment.GET("/suggestions", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.ListReorderSuggestions)
			replenishment.POST("/suggestions/:suggestionId/create-po", middleware.RequirePermission(roleRepo, "replenishment.write"), handlers.CreatePOFromSuggestion)
			replenishment.POST("/purchase-orders/:poId/approve", middleware.RequirePermission(roleRepo, "replenishment.write"), handlers.ApprovePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/send", middleware.RequirePermission(roleRepo, "replenishment.write"), handlers.SendPurchaseOrder)
			replenishment.GET("/purchase-orders/:poId", middleware.RequirePermission(roleRepo, "replenishment.read"), handlers.GetPurchaseOrder)
			replenishment.PUT("/purchase-orders/:poId", middleware.RequirePermission(roleRepo, "replenishment.write"), handlers.UpdatePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/receive", middleware.RequirePermission(roleRepo, "replenishment.write"), handlers.ReceivePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/cancel", middleware.RequirePermission(roleRepo, "replenishment.write"), handlers.CancelPurchaseOrder)
		}

		// Sales
		sales := api.Group("/sales")
		{
			sales.POST("/checkout", middleware.RequirePermission(roleRepo, "pos.access"), handlers.NewSalesHandler(db, settingsService).Checkout)
			sales.GET("/products", middleware.RequirePermission(roleRepo, "pos.access"), handlers.NewSalesHandler(db, settingsService).ListProducts)
			sales.GET("/orders", middleware.RequirePermission(roleRepo, "pos.access"), handlers.NewSalesHandler(db, settingsService).ListOrders)
		}

		// Returns
		returns := api.Group("/returns")
		{
			returns.POST("/request", middleware.RequirePermission(roleRepo, "returns.request"), returnHandler.RequestReturn)
			returns.POST("/:id/process", middleware.RequirePermission(roleRepo, "returns.manage"), returnHandler.ProcessReturn)
			returns.GET("", middleware.RequirePermission(roleRepo, "returns.manage"), returnHandler.ListReturns)
		}

		// Reports
		reports := api.Group("/reports")
		{
			reports.POST("/sales-trends", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetSalesTrendsReport)
			reports.POST("/sales-trends/export", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.ExportSalesTrendsReport)
			reports.GET("/jobs/:jobId", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetReportJobStatus)
			reports.GET("/download/:jobId", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.DownloadReportFile)
			reports.POST("/inventory-turnover", middleware.RequirePermission(roleRepo, "reports.inventory"), reportHandler.GetInventoryTurnoverReport)
			reports.POST("/profit-margin", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetProfitMarginReport)
		}

		// Alerts
		alerts := api.Group("/alerts")
		{
			alerts.GET("", middleware.RequirePermission(roleRepo, "alerts.view"), handlers.ListAlerts)
			alerts.GET("/:alertId", middleware.RequirePermission(roleRepo, "alerts.view"), handlers.GetAlert)
			alerts.PATCH("/:alertId/resolve", middleware.RequirePermission(roleRepo, "alerts.manage"), handlers.ResolveAlert)
			alerts.PUT("/products/:productId/settings", middleware.RequirePermission(roleRepo, "alerts.manage"), handlers.PutProductAlertSettings)
			alerts.PUT("/users/:userId/notification-settings", middleware.RequirePermission(roleRepo, "settings.manage"), handlers.PutUserNotificationSettings)
			alerts.POST("/check", func(c *gin.Context) {
				handlers.CheckAndTriggerAlerts()
				c.JSON(http.StatusOK, gin.H{"message": "Alert check triggered"})
			})
			// Subscriptions - Admin Only
			subscriptions := alerts.Group("/subscriptions")
			subscriptions.Use(middleware.AdminOnly())
			{
				subscriptions.POST("", handlers.CreateAlertRoleSubscription)
				subscriptions.GET("", handlers.ListAlertRoleSubscriptions)
				subscriptions.DELETE("/:id", handlers.DeleteAlertRoleSubscription)
			}
		}

		// Bulk Operations
		bulk := api.Group("/bulk")
		{
			bulk.GET("/products/template", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.GetProductImportTemplate)
			bulk.POST("/products/import", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.UploadProductImport)
			bulk.GET("/products/import/:jobId/status", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.GetBulkImportStatus)
			bulk.POST("/products/import/:jobId/confirm", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.ConfirmBulkImport)
			bulk.GET("/products/export", middleware.RequirePermission(roleRepo, "bulk.export"), bulkHandler.ExportProducts)
			bulk.GET("/jobs", middleware.RequirePermission(roleRepo, "bulk.import"), bulkHandler.ListBulkJobs)
			bulk.GET("/files/:bucket/:object", middleware.RequirePermission(roleRepo, "bulk.export"), bulkHandler.DownloadFile)
		}

		// Inventory (Transfers)
		inventory := api.Group("/inventory")
		{
			inventory.POST("/transfers", middleware.RequirePermission(roleRepo, "products.write"), handlers.CreateStockTransfer)
		}

		// Users - Protected mainly by users.manage
		users := api.Group("/users")
		{
			users.GET("", middleware.RequirePermission(roleRepo, "users.view"), userHandler.ListUsers)
			users.POST("/refresh-token", userHandler.RefreshToken)
			users.POST("/logout", userHandler.LogoutUser)
			users.GET("/:id", userHandler.GetUser) // Might need self-access check
			users.PUT("/:id", middleware.RequirePermission(roleRepo, "users.manage"), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequirePermission(roleRepo, "users.manage"), userHandler.DeleteUser)
			users.PUT("/:id/approve", middleware.RequirePermission(roleRepo, "users.manage"), userHandler.ApproveUser)

			// Notification routes (Self access usually)
			notifications := users.Group("/:id/notifications")
			{
				notifications.GET("", notificationHandler.ListNotifications)
				notifications.GET("/unread/count", notificationHandler.GetUnreadNotificationCount)
				notifications.PATCH("/:notificationId/read", notificationHandler.MarkNotificationAsRead)
				notifications.PATCH("/read-all", notificationHandler.MarkAllNotificationsAsRead)
			}
		}

		// CRM
		crm := api.Group("/crm")
		{
			crm.GET("/customers", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.ListCustomers)
			crm.POST("/customers", middleware.RequirePermission(roleRepo, "customers.write"), crmHandler.CreateCustomer)
			crm.GET("/customers/:identifier", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomer)
			crm.GET("/customers/username/:username", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomerByUsername)
			crm.GET("/customers/email/:email", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomerByEmail)
			crm.GET("/customers/phone/:phone", middleware.RequirePermission(roleRepo, "customers.read"), crmHandler.GetCustomerByPhone)
			crm.PUT("/customers/:userId", middleware.RequirePermission(roleRepo, "customers.write"), crmHandler.UpdateCustomer)
			crm.DELETE("/customers/:userId", middleware.RequirePermission(roleRepo, "customers.write"), crmHandler.DeleteCustomer)
			crm.GET("/loyalty/:userId", middleware.RequirePermission(roleRepo, "loyalty.read"), crmHandler.GetLoyaltyAccount)
			crm.POST("/loyalty/:userId/points", middleware.RequirePermission(roleRepo, "loyalty.write"), crmHandler.AddLoyaltyPoints)
			crm.POST("/loyalty/:userId/redeem", middleware.RequirePermission(roleRepo, "loyalty.write"), crmHandler.RedeemLoyaltyPoints)
		}

		// Time Tracking (Basic access for all staff)
		timeTracking := api.Group("/time-tracking")
		{
			// All logged in users can clock in? Or need permission?
			// Let's assume basic staff role has access, or implicit if authenticated.
			// But for strict RBAC, let's use a generic permission or keep basic auth?
			// The previous code had AuthorizeMiddleware("Admin", "Manager", "Staff").
			// I'll add "timeclock.use" or just allow all authenticated for now?
			// No, let's use "pos.access" or create "timeclock.use"?
			// I didn't seed "timeclock.use".
			// I'll use "pos.access" for now as it implies staff presence, OR create seeded permission?
			// Creating new seed is safer.
			// For now, I'll rely on roleRepo check (if user has role). "Staff" role has it.
			// If I use "pos.access", manager has it.
			// I'll stick to NO specific permission for time tracking self-actions, just Auth.
			// But wait, "GetTeamOverview" needs Manager access.
			timeTracking.POST("/clock-in/:userId", timeTrackingHandler.ClockIn)
			timeTracking.POST("/clock-out/:userId", timeTrackingHandler.ClockOut)
			timeTracking.POST("/break-start/:userId", timeTrackingHandler.StartBreak)
			timeTracking.POST("/break-end/:userId", timeTrackingHandler.EndBreak)
			timeTracking.GET("/last-entry/:userId", timeTrackingHandler.GetLastTimeClock)
			timeTracking.GET("/history/:userId", timeTrackingHandler.GetHistory)
			timeTracking.GET("/activities", timeTrackingHandler.GetRecentActivities)
			timeTracking.GET("/weekly-summary/:userId", timeTrackingHandler.GetWeeklySummary)
			// Manager/Admin only
			timeTracking.GET("/team-status", middleware.RequirePermission(roleRepo, "users.manage"), timeTrackingHandler.GetTeamStatus)
			timeTracking.GET("/team-overview", middleware.RequirePermission(roleRepo, "users.manage"), timeTrackingHandler.GetTeamOverview)
		}

		// Global Search
		api.GET("/search", searchHandler.Search) // Open search? Or inventory.read? Let's keep open for authenticated.

		// Settings
		settings := api.Group("/settings")
		{
			settings.GET("", middleware.RequirePermission(roleRepo, "settings.view"), settingsHandler.GetSettings)
			settings.PUT("", middleware.RequirePermission(roleRepo, "settings.manage"), settingsHandler.UpdateSetting)
		}

		// Roles
		roles := api.Group("/roles")
		{
			roles.GET("", middleware.RequirePermission(roleRepo, "roles.view"), roleHandler.ListRoles)
			roles.POST("", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.CreateRole)
			roles.PUT("/:id", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.UpdateRole)
			roles.DELETE("/:id", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.DeleteRole)
			roles.PUT("/:id/permissions", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.UpdateRolePermissions)
		}

		permissions := api.Group("/permissions")
		{
			permissions.GET("", middleware.RequirePermission(roleRepo, "roles.view"), roleHandler.ListPermissions)
		}
	}

	return r
}
