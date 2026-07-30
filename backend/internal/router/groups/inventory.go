package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterInventory(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	inventory := rg.Group("/inventory")
	{
		inventory.POST("/transfers", middleware.RequirePermission(roleRepo, "products.write"), handlers.CreateStockTransfer)
	}
}
