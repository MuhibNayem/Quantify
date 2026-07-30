package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterSuppliers(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	supplierHandler := c.Supplier

	suppliers := rg.Group("/suppliers")
	{
		suppliers.POST("", middleware.RequirePermission(roleRepo, "suppliers.write"), supplierHandler.CreateSupplier)
		suppliers.GET("", middleware.RequirePermission(roleRepo, "suppliers.read"), supplierHandler.ListSuppliers)
		suppliers.GET("/:id", middleware.RequirePermission(roleRepo, "suppliers.read"), supplierHandler.GetSupplier)
		suppliers.GET("/name/:name", middleware.RequirePermission(roleRepo, "suppliers.read"), supplierHandler.GetSupplierByName)
		suppliers.PUT("/:id", middleware.RequirePermission(roleRepo, "suppliers.write"), supplierHandler.UpdateSupplier)
		suppliers.DELETE("/:id", middleware.RequirePermission(roleRepo, "suppliers.write"), supplierHandler.DeleteSupplier)
		suppliers.GET("/:id/performance", middleware.RequirePermission(roleRepo, "reports.financial"), supplierHandler.GetSupplierPerformanceReport)
	}
}
