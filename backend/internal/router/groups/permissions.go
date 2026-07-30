package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterPermissions(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	roleHandler := c.RoleHandler

	permissions := rg.Group("/permissions")
	{
		permissions.GET("", middleware.RequirePermission(roleRepo, "roles.view"), roleHandler.ListPermissions)
	}
}
