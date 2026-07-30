package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterReturns(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	returnHandler := c.Return

	returns := rg.Group("/returns")
	{
		returns.POST("/request", middleware.RequirePermission(roleRepo, "returns.request"), returnHandler.RequestReturn)
		returns.POST("/:id/process", middleware.RequirePermission(roleRepo, "returns.manage"), returnHandler.ProcessReturn)
		returns.GET("", middleware.RequirePermission(roleRepo, "returns.manage"), returnHandler.ListReturns)
		returns.GET("/:id", middleware.RequirePermission(roleRepo, "returns.manage"), returnHandler.GetReturn)
	}
}
