package groups

import (
	"net/http"

	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterAlerts(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	alerts := rg.Group("/alerts")
	{
		alerts.GET("", middleware.RequirePermission(roleRepo, "alerts.view"), handlers.ListAlerts)
		alerts.GET("/:alertId", middleware.RequirePermission(roleRepo, "alerts.view"), handlers.GetAlert)
		alerts.PATCH("/:alertId/resolve", middleware.RequirePermission(roleRepo, "alerts.manage"), handlers.ResolveAlert)
		alerts.PUT("/products/:productId/settings", middleware.RequirePermission(roleRepo, "alerts.manage"), handlers.PutProductAlertSettings)
		alerts.PUT("/users/:userId/notification-settings", middleware.RequirePermission(roleRepo, "settings.manage"), handlers.PutUserNotificationSettings)
		alerts.POST("/check", func(c *gin.Context) {
			handlers.CheckAndTriggerAlerts()
			c.JSON(http.StatusOK, gin.H{"message": "Alert check triggered"})
		})
		subscriptions := alerts.Group("/subscriptions")
		subscriptions.Use(middleware.AdminOnly())
		{
			subscriptions.POST("", handlers.CreateAlertRoleSubscription)
			subscriptions.GET("", handlers.ListAlertRoleSubscriptions)
			subscriptions.DELETE("/:id", handlers.DeleteAlertRoleSubscription)
		}
	}
}
