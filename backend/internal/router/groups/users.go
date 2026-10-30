package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterUsers(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	userHandler := c.User
	notificationHandler := c.Notification

	users := rg.Group("/users")
	{
		users.GET("", middleware.RequirePermission(roleRepo, "users.view"), userHandler.ListUsers)
		users.GET("/me", userHandler.GetCurrentUser)
		users.GET("/:id", userHandler.GetUser)
		users.PUT("/:id", middleware.RequirePermission(roleRepo, "users.manage"), userHandler.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission(roleRepo, "users.manage"), userHandler.DeleteUser)
		users.PUT("/:id/approve", middleware.RequirePermission(roleRepo, "users.manage"), userHandler.ApproveUser)

		notifications := users.Group("/:id/notifications")
		{
			notifications.GET("", middleware.RequirePermission(roleRepo, "notifications.read"), notificationHandler.ListNotifications)
			notifications.GET("/unread/count", middleware.RequirePermission(roleRepo, "notifications.read"), notificationHandler.GetUnreadNotificationCount)
			notifications.PATCH("/:notificationId/read", middleware.RequirePermission(roleRepo, "notifications.write"), notificationHandler.MarkNotificationAsRead)
			notifications.PATCH("/read-all", middleware.RequirePermission(roleRepo, "notifications.write"), notificationHandler.MarkAllNotificationsAsRead)
			notifications.PUT("/subscriptions/:type", middleware.RequirePermission(roleRepo, "notifications.write"), notificationHandler.UpdateSubscription)
			notifications.GET("/subscriptions", middleware.RequirePermission(roleRepo, "notifications.read"), notificationHandler.ListSubscriptions)
		}
	}
}
