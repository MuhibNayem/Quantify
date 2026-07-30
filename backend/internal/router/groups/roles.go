package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterRoles(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	roleHandler := c.RoleHandler

	roles := rg.Group("/roles")
	{
		roles.GET("", middleware.RequirePermission(roleRepo, "roles.view"), roleHandler.ListRoles)
		roles.POST("", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.CreateRole)
		roles.PUT("/:id", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.UpdateRole)
		roles.DELETE("/:id", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.DeleteRole)
		roles.PUT("/:id/permissions", middleware.RequirePermission(roleRepo, "roles.manage"), roleHandler.UpdateRolePermissions)
	}
}
