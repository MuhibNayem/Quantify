package router

import (
	"net/http"

	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "inventory/backend/api"
)

func SetupRouter(hub *websocket.Hub) *gin.Engine {
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		MaxAge:           12 * 3600,
		AllowCredentials: true,
	}))

	r.Use(middleware.ErrorHandler())

	// Public routes (no tenant middleware)
	publicRoutes := r.Group("/")
	{
		publicRoutes.GET("/health", handlers.HealthCheck)
		publicRoutes.GET("/ws", func(c *gin.Context) {
			handlers.ServeWs(hub, c)
		})
		publicRoutes.GET("/metrics", gin.WrapH(promhttp.Handler()))
		// Swagger UI
		publicRoutes.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Public API routes (no auth middleware)
	publicAPI := r.Group("/api/v1")
	{
		userRoutes := publicAPI.Group("/users")
		{
			userRoutes.POST("/register", handlers.RegisterUser)
			userRoutes.POST("/login", handlers.LoginUser)
		}
	}

	api := r.Group("/api/v1")
	{
		// Auth middleware will be applied to all routes in this group
		api.Use(middleware.AuthMiddleware())

		// Products
		products := api.Group("/products")
		{
			products.POST("", handlers.CreateProduct)
			products.GET("", handlers.ListProducts)
			products.GET("/:id", handlers.GetProduct)
			products.PUT("/:id", handlers.UpdateProduct)
			products.DELETE("/:id", handlers.DeleteProduct)
			products.GET("/:id/stock", handlers.GetProductStock)
			products.POST("/:productId/stock/batches", handlers.CreateBatch)
			products.POST("/:productId/stock/adjustments", handlers.CreateStockAdjustment)
			products.GET("/:id/history", handlers.ListStockHistory)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.POST("", handlers.CreateCategory)
			categories.GET("", handlers.ListCategories)
			categories.GET("/:id", handlers.GetCategory)
			categories.PUT("/:id", handlers.UpdateCategory)
			categories.DELETE("/:id", handlers.DeleteCategory)
		}

		// Sub-categories
		subCategories := api.Group("/sub-categories")
		{
			subCategories.POST("", handlers.CreateSubCategory)
			subCategories.GET("", handlers.ListSubCategories)
			subCategories.PUT("/:id", handlers.UpdateSubCategory)
			subCategories.DELETE("/:id", handlers.DeleteSubCategory)
		}

		// Suppliers
		suppliers := api.Group("/suppliers")
		{
			suppliers.POST("", handlers.CreateSupplier)
			suppliers.GET("", handlers.ListSuppliers)
			suppliers.GET("/:id", handlers.GetSupplier)
			suppliers.PUT("/:id", handlers.UpdateSupplier)
			suppliers.DELETE("/:id", handlers.DeleteSupplier)
			suppliers.GET("/:id/performance", handlers.GetSupplierPerformanceReport)
		}

		// Locations
		locations := api.Group("/locations")
		{
			locations.POST("", handlers.CreateLocation)
			locations.GET("", handlers.ListLocations)
			locations.GET("/:id", handlers.GetLocation)
			locations.PUT("/:id", handlers.UpdateLocation)
			locations.DELETE("/:id", handlers.DeleteLocation)
		}

		// Barcode
		barcode := api.Group("/barcode")
		{
			barcode.GET("/lookup", handlers.LookupProductByBarcode)
			barcode.GET("/generate", handlers.GenerateBarcode)
		}

		// Replenishment
		replenishment := api.Group("/replenishment")
		{
			replenishment.POST("/forecast/generate", handlers.GenerateDemandForecast)
			replenishment.GET("/forecast/:forecastId", handlers.GetDemandForecast)
			replenishment.GET("/suggestions", handlers.ListReorderSuggestions)
			replenishment.POST("/suggestions/:suggestionId/create-po", handlers.CreatePOFromSuggestion)
			replenishment.POST("/purchase-orders/:poId/approve", handlers.ApprovePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/send", handlers.SendPurchaseOrder)
			replenishment.GET("/purchase-orders/:poId", handlers.GetPurchaseOrder)
			replenishment.PUT("/purchase-orders/:poId", handlers.UpdatePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/receive", handlers.ReceivePurchaseOrder)
			replenishment.POST("/purchase-orders/:poId/cancel", handlers.CancelPurchaseOrder)
		}

		// Reports
		reports := api.Group("/reports")
		{
			reports.POST("/sales-trends", handlers.GetSalesTrendsReport)
			reports.POST("/inventory-turnover", handlers.GetInventoryTurnoverReport)
			reports.POST("/profit-margin", handlers.GetProfitMarginReport)
		}

		// Alerts
		alerts := api.Group("/alerts")
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
		}

		// Bulk Operations
		bulk := api.Group("/bulk")
		{
			bulk.GET("/products/template", handlers.GetProductImportTemplate)
			bulk.POST("/products/import", handlers.UploadProductImport)
			bulk.GET("/products/import/:jobId/status", handlers.GetBulkImportStatus)
			bulk.POST("/products/import/:jobId/confirm", handlers.ConfirmBulkImport)
			bulk.GET("/products/export", handlers.ExportProducts)
		}

		// Inventory
		inventory := api.Group("/inventory")
		{
			inventory.POST("/transfers", handlers.CreateStockTransfer)
		}

		// Users
		users := api.Group("/users")
		{
			users.GET("", handlers.ListUsers)
			users.POST("/refresh-token", handlers.RefreshToken)
			users.POST("/logout", handlers.LogoutUser)
			users.GET("/:id", handlers.GetUser)
			users.PUT("/:id", middleware.AdminOnly(), handlers.UpdateUser)
			users.DELETE("/:id", middleware.AdminOnly(), handlers.DeleteUser)
			users.PUT("/:id/approve", middleware.AdminOnly(), handlers.ApproveUser)
		}
	}

	return r
}
