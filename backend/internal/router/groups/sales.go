package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterSales(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	salesHandler := c.SalesHandler

	sales := rg.Group("/sales")
	{
		sales.POST("/checkout", middleware.RequirePermission(roleRepo, "pos.access"), salesHandler.Checkout)
		sales.GET("/products", middleware.RequirePermission(roleRepo, "pos.access"), salesHandler.ListProducts)
		sales.GET("/orders", middleware.RequirePermission(roleRepo, "pos.view"), salesHandler.ListOrders)
		sales.GET("/orders/:orderNumber", middleware.RequirePermission(roleRepo, "pos.view"), salesHandler.GetOrderByNumber)
		sales.GET("/history", middleware.RequirePermission(roleRepo, "pos.view"), salesHandler.ListAllOrders)
	}
}
