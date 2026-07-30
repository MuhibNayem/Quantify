package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterProducts(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	productHandler := c.Product
	unitsHandler := handlers.NewUnitsHandler(c.DB)

	products := rg.Group("/products")
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
		products.GET("/:productId/units", middleware.RequirePermission(roleRepo, "products.read"), unitsHandler.ListProductUnits)
		products.POST("/:productId/units", middleware.RequirePermission(roleRepo, "products.write"), unitsHandler.SaveProductUnit)
		products.DELETE("/:productId/units/:unitId", middleware.RequirePermission(roleRepo, "products.write"), unitsHandler.DeleteProductUnit)
	}
}
