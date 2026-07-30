package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterNotifications(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	notificationHandler := c.Notification

	notifications := rg.Group("/notifications")
	{
		notifications.POST("/broadcast", middleware.RequirePermission(roleRepo, "notifications.write"), notificationHandler.BroadcastNotification)
	}
}
