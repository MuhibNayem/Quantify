package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterCashDrawer(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	handler := handlers.NewCashDrawerHandler(c.DB)
	drawer := rg.Group("/pos/cash-drawer")
	drawer.Use(middleware.RequirePermission(roleRepo, "pos.access"))
	{
		drawer.GET("/current", handler.Current)
		drawer.GET("/sessions", handler.List)
		drawer.POST("/open", handler.Open)
		drawer.POST("/drop", handler.Drop)
		drawer.POST("/close", handler.Close)
	}
}
