package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterLocations(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	locations := rg.Group("/locations")
	{
		locations.POST("", middleware.RequirePermission(roleRepo, "locations.write"), handlers.CreateLocation)
		locations.GET("", middleware.RequirePermission(roleRepo, "locations.read"), handlers.ListLocations)
		locations.GET("/:id", middleware.RequirePermission(roleRepo, "locations.read"), handlers.GetLocation)
		locations.PUT("/:id", middleware.RequirePermission(roleRepo, "locations.write"), handlers.UpdateLocation)
		locations.DELETE("/:id", middleware.RequirePermission(roleRepo, "locations.write"), handlers.DeleteLocation)
	}
}
