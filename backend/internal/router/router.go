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

	// Initialize services
	paymentService := services.NewPaymentService(cfg, paymentRepo)
	forecastingService := services.NewForecastingService(forecastingRepo)
	barcodeService := services.NewBarcodeService(barcodeRepo)
	crmService := services.NewCRMService(crmRepo)
	timeTrackingService := services.NewTimeTrackingService(timeTrackingRepo)
	integrationService := services.NewIntegrationService()
	reportingService := services.NewReportingService(reportsRepo, minioUploader, jobRepo, cfg)
	searchService := services.NewSearchService(db, searchRepo, productRepo, userRepo, supplierRepo, categoryRepo)
	replenishmentService := services.NewReplenishmentService(replenishmentRepo)

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
		api.GET("/dashboard/summary", dashboardHandler.GetDashboardSummary)

		// Products
		products := api.Group("/products")
		{
			products.POST("", middleware.AuthorizeMiddleware("Admin", "Manager"), productHandler.CreateProduct)
			products.GET("", productHandler.ListProducts)
			products.GET("/:productId", productHandler.GetProduct)
			products.GET("/sku/:sku", productHandler.GetProductBySKU)
			products.GET("/barcode/:barcode", productHandler.GetProductByBarcode)
			products.PUT("/:productId", middleware.AuthorizeMiddleware("Admin", "Manager"), productHandler.UpdateProduct)
			products.DELETE("/:productId", middleware.AuthorizeMiddleware("Admin", "Manager"), productHandler.DeleteProduct)
			products.GET("/:productId/stock", handlers.GetProductStock)
			products.POST("/:productId/stock/batches", middleware.AuthorizeMiddleware("Admin", "Manager"), handlers.CreateBatch)
			products.POST("/:productId/stock/adjustments", middleware.AuthorizeMiddleware("Admin", "Manager", "Staff"), handlers.CreateStockAdjustment)
			products.GET("/:productId/history", handlers.ListStockHistory)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.POST("", middleware.AuthorizeMiddleware("Admin", "Manager"), categoryHandler.CreateCategory)
			categories.GET("", categoryHandler.ListCategories)
			categories.GET("/:categoryId", categoryHandler.GetCategory)
			categories.GET("/name/:name", categoryHandler.GetCategoryByName)
			categories.PUT("/:categoryId", middleware.AuthorizeMiddleware("Admin", "Manager"), categoryHandler.UpdateCategory)
			categories.DELETE("/:categoryId", middleware.AuthorizeMiddleware("Admin", "Manager"), categoryHandler.DeleteCategory)

			categories.POST("/:categoryId/sub-categories", middleware.AuthorizeMiddleware("Admin", "Manager"), categoryHandler.CreateSubCategory)
			categories.GET("/:categoryId/sub-categories", categoryHandler.ListSubCategories)
		}

		// Sub-categories
		subCategories := api.Group("/sub-categories")
		{
			subCategories.PUT("/:id", middleware.AuthorizeMiddleware("Admin", "Manager"), categoryHandler.UpdateSubCategory)
			subCategories.DELETE("/:id", middleware.AuthorizeMiddleware("Admin", "Manager"), categoryHandler.DeleteSubCategory)
		}

		// Suppliers
		suppliers := api.Group("/suppliers")
		{
			suppliers.POST("", middleware.AuthorizeMiddleware("Admin", "Manager"), supplierHandler.CreateSupplier)
			suppliers.GET("", supplierHandler.ListSuppliers)
			suppliers.GET("/:id", supplierHandler.GetSupplier)
			suppliers.GET("/name/:name", supplierHandler.GetSupplierByName)
			suppliers.PUT("/:id", middleware.AuthorizeMiddleware("Admin", "Manager"), supplierHandler.UpdateSupplier)
			suppliers.DELETE("/:id", middleware.AuthorizeMiddleware("Admin", "Manager"), supplierHandler.DeleteSupplier)
			suppliers.GET("/:id/performance", supplierHandler.GetSupplierPerformanceReport)
		}

		// Locations
		locations := api.Group("/locations")
		{
			locations.POST("", middleware.AuthorizeMiddleware("Admin", "Manager"), handlers.CreateLocation)
			locations.GET("", handlers.ListLocations)
			locations.GET("/:id", handlers.GetLocation)
			locations.PUT("/:id", middleware.AuthorizeMiddleware("Admin", "Manager"), handlers.UpdateLocation)
			locations.DELETE("/:id", middleware.AuthorizeMiddleware("Admin", "Manager"), handlers.DeleteLocation)
		}

		// Barcode
		barcode := api.Group("/barcode")
		{
			barcode.GET("/lookup", barcodeHandler.LookupProductByBarcode)
			barcode.GET("/generate", barcodeHandler.GenerateBarcode)
		}

		// Replenishment (Manager/Admin)
		replenishment := api.Group("/replenishment")
		replenishment.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
		{
			replenishment.POST("/forecast/generate", replenishmentHandler.GenerateDemandForecast)
			replenishment.POST("/suggestions/generate", replenishmentHandler.GenerateReorderSuggestions)
			replenishment.GET("/forecast/:forecastId", handlers.GetDemandForecast)
			replenishment.GET("/suggestions", replenishmentHandler.ListReorderSuggestions)
			replenishment.POST("/suggestions/:suggestionId/create-po", handlers.CreatePOFromSuggestion)
			replenishment.POST("/purchase-orders/:poId/approve", handlers.ApprovePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/send", handlers.SendPurchaseOrder)
			replenishment.GET("/purchase-orders/:poId", handlers.GetPurchaseOrder)
			replenishment.PUT("/purchase-orders/:poId", handlers.UpdatePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/receive", handlers.ReceivePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/cancel", handlers.CancelPurchaseOrder)
		}

		// Sales
		sales := api.Group("/sales")
		sales.Use(middleware.AuthorizeMiddleware("Admin", "Manager", "Staff"))
		{
			sales.POST("/checkout", handlers.NewSalesHandler(db).Checkout)
			sales.GET("/products", handlers.NewSalesHandler(db).ListProducts)
		}

		// Reports (Manager/Admin)
		reports := api.Group("/reports")
		reports.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
		{
			reports.POST("/sales-trends", reportHandler.GetSalesTrendsReport)
			reports.POST("/sales-trends/export", reportHandler.ExportSalesTrendsReport)
			reports.GET("/jobs/:jobId", reportHandler.GetReportJobStatus)
			reports.GET("/download/:jobId", reportHandler.DownloadReportFile)
			reports.POST("/inventory-turnover", reportHandler.GetInventoryTurnoverReport)
			reports.POST("/profit-margin", reportHandler.GetProfitMarginReport)
		}

		// Alerts (Manager/Admin)
		alerts := api.Group("/alerts")
		alerts.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
		{
			alerts.GET("", handlers.ListAlerts)
			alerts.GET("/:alertId", handlers.GetAlert)
			alerts.PATCH("/:alertId/resolve", handlers.ResolveAlert)
			alerts.PUT("/products/:productId/settings", handlers.PutProductAlertSettings)
			alerts.PUT("/users/:userId/notification-settings", handlers.PutUserNotificationSettings)
			alerts.POST("/check", func(c *gin.Context) {
				handlers.CheckAndTriggerAlerts()
				c.JSON(http.StatusOK, gin.H{"message": "Alert check triggered"})
			})
			// Alert Role Subscriptions (Admin Only)
			subscriptions := alerts.Group("/subscriptions")
			subscriptions.Use(middleware.AdminOnly())
			{
				subscriptions.POST("", handlers.CreateAlertRoleSubscription)
				subscriptions.GET("", handlers.ListAlertRoleSubscriptions)
				subscriptions.DELETE("/:id", handlers.DeleteAlertRoleSubscription)
			}
		}

		// Bulk Operations (Manager/Admin)
		bulk := api.Group("/bulk")
		bulk.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
		{
			bulk.GET("/products/template", bulkHandler.GetProductImportTemplate)
			bulk.POST("/products/import", bulkHandler.UploadProductImport)
			bulk.GET("/products/import/:jobId/status", bulkHandler.GetBulkImportStatus)
			bulk.POST("/products/import/:jobId/confirm", bulkHandler.ConfirmBulkImport)
			bulk.GET("/products/export", bulkHandler.ExportProducts)
			bulk.GET("/jobs", bulkHandler.ListBulkJobs)
			bulk.GET("/files/:bucket/:object", bulkHandler.DownloadFile)
		}

		// Inventory (Staff and above)
		inventory := api.Group("/inventory")
		inventory.Use(middleware.AuthorizeMiddleware("Admin", "Manager", "Staff"))
		{
			inventory.POST("/transfers", handlers.CreateStockTransfer)
		}

		// Users
		users := api.Group("/users")
		{
			users.GET("", middleware.AdminOnly(), userHandler.ListUsers)
			users.POST("/refresh-token", userHandler.RefreshToken)
			users.POST("/logout", userHandler.LogoutUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", middleware.AdminOnly(), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.AdminOnly(), userHandler.DeleteUser)
			users.PUT("/:id/approve", middleware.AdminOnly(), userHandler.ApproveUser)

			// Notification routes
			notifications := users.Group("/:id/notifications")
			{
				notifications.GET("", notificationHandler.ListNotifications)
				notifications.GET("/unread/count", notificationHandler.GetUnreadNotificationCount)
				notifications.PATCH("/:notificationId/read", notificationHandler.MarkNotificationAsRead)
				notifications.PATCH("/read-all", notificationHandler.MarkAllNotificationsAsRead)
			}
		}

		// CRM (Manager/Admin)
		crm := api.Group("/crm")
		crm.Use(middleware.AuthorizeMiddleware("Admin", "Manager"))
		{
			crm.POST("/customers", crmHandler.CreateCustomer)
			crm.GET("/customers/:identifier", crmHandler.GetCustomer)
			crm.GET("/customers/username/:username", crmHandler.GetCustomerByUsername)
			crm.GET("/customers/email/:email", crmHandler.GetCustomerByEmail)
			crm.GET("/customers/phone/:phone", crmHandler.GetCustomerByPhone)
			crm.PUT("/customers/:userId", crmHandler.UpdateCustomer)
			crm.DELETE("/customers/:userId", crmHandler.DeleteCustomer)
			crm.GET("/loyalty/:userId", crmHandler.GetLoyaltyAccount)
			crm.POST("/loyalty/:userId/points", crmHandler.AddLoyaltyPoints)
			crm.POST("/loyalty/:userId/redeem", crmHandler.RedeemLoyaltyPoints)
		}

		// Time Tracking (Staff and above)
		timeTracking := api.Group("/time-tracking")
		timeTracking.Use(middleware.AuthorizeMiddleware("Admin", "Manager", "Staff"))
		{
			timeTracking.POST("/clock-in/:userId", timeTrackingHandler.ClockIn)
			timeTracking.POST("/clock-out/:userId", timeTrackingHandler.ClockOut)
			timeTracking.GET("/last-entry/:userId", timeTrackingHandler.GetLastTimeClock)
			timeTracking.GET("/last-entry/username/:username", timeTrackingHandler.GetLastTimeClockByUsername)
		}
		// Global Search
		api.GET("/search", searchHandler.Search)
	}

	return r
}
