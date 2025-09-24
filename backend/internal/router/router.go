package router

import (
	"net/http"

	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/websocket"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "inventory/backend/api"
)

func SetupRouter(hub *websocket.Hub) *gin.Engine {
	r := gin.Default()

	// pprof
	pprof.Register(r)

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

	// Routes requiring tenant middleware
	tenantRoutes := r.Group("/")
	tenantRoutes.Use(middleware.TenantMiddleware())

	api := tenantRoutes.Group("/api/v1")
	{
		// Auth middleware will be applied to all routes in this group
		api.Use(middleware.AuthMiddleware())

		// Products
		products := api.Group("/products")
		{
			products.POST("/", handlers.CreateProduct)
			products.GET("/", handlers.ListProducts)
			products.GET("/:id", handlers.GetProduct)
			products.PUT("/:id", handlers.UpdateProduct)
			products.DELETE("/:id", handlers.DeleteProduct)
			products.GET("/:id/stock", handlers.GetProductStock)
			products.GET("/:id/history", handlers.ListStockHistory)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.POST("/", handlers.CreateCategory)
			categories.GET("/", handlers.ListCategories)
			categories.GET("/:id", handlers.GetCategory)
			categories.PUT("/:id", handlers.UpdateCategory)
			categories.DELETE("/:id", handlers.DeleteCategory)
		}

		// Suppliers
		suppliers := api.Group("/suppliers")
		{
			suppliers.POST("/", handlers.CreateSupplier)
			suppliers.GET("/", handlers.ListSuppliers)
			suppliers.GET("/:id", handlers.GetSupplier)
			suppliers.PUT("/:id", handlers.UpdateSupplier)
			suppliers.DELETE("/:id", handlers.DeleteSupplier)
			suppliers.GET("/:id/performance", handlers.GetSupplierPerformanceReport)
		}

		// Locations
		locations := api.Group("/locations")
		{
			locations.POST("/", handlers.CreateLocation)
			locations.GET("/", handlers.ListLocations)
			locations.GET("/:id", handlers.GetLocation)
			locations.PUT("/:id", handlers.UpdateLocation)
			locations.DELETE("/:id", handlers.DeleteLocation)
		}

		// Stock
		stock := api.Group("/stock")
		{
			stock.POST("/adjustments", handlers.CreateStockAdjustment)
		}

		// Barcode
		barcode := api.Group("/barcode")
		{
			barcode.GET("/lookup", handlers.LookupProductByBarcode)
			barcode.POST("/generate", handlers.GenerateBarcode)
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
			users.POST("/register", handlers.RegisterUser)
			users.POST("/login", handlers.LoginUser)
			users.POST("/refresh-token", handlers.RefreshToken)
			users.POST("/logout", handlers.LogoutUser)
			users.GET("/:id", handlers.GetUser)
			users.PUT("/:id", handlers.UpdateUser)
			users.DELETE("/:id", handlers.DeleteUser)
		}
	}

	return r
}
