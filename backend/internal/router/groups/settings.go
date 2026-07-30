package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterSettings(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	settingsHandler := c.SettingsHandler

	settings := rg.Group("/settings")
	{
		settings.GET("", middleware.RequirePermission(roleRepo, "settings.view"), settingsHandler.GetSettings)
		settings.GET("/configurations", settingsHandler.GetPublicConfigurations)
		settings.PUT("", middleware.RequirePermission(roleRepo, "settings.manage"), settingsHandler.UpdateSetting)
	}
}
