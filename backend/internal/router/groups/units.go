package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterUnits(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	unitsHandler := handlers.NewUnitsHandler(c.DB)

	units := rg.Group("/units")
	{
		units.GET("", middleware.RequirePermission(roleRepo, "products.read"), unitsHandler.ListUnits)
		units.GET("/dimensions", middleware.RequirePermission(roleRepo, "products.read"), unitsHandler.ListDimensions)
	}
}
