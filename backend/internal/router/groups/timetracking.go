package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterTimeTracking(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	timeTrackingHandler := c.TimeTrackingHandler

	timeTracking := rg.Group("/time-tracking")
	{
		timeTracking.POST("/clock-in/:userId", timeTrackingHandler.ClockIn)
		timeTracking.POST("/clock-out/:userId", timeTrackingHandler.ClockOut)
		timeTracking.POST("/break-start/:userId", timeTrackingHandler.StartBreak)
		timeTracking.POST("/break-end/:userId", timeTrackingHandler.EndBreak)
		timeTracking.GET("/last-entry/:userId", timeTrackingHandler.GetLastTimeClock)
		timeTracking.GET("/history/:userId", timeTrackingHandler.GetHistory)
		timeTracking.GET("/activities", timeTrackingHandler.GetRecentActivities)
		timeTracking.GET("/weekly-summary/:userId", timeTrackingHandler.GetWeeklySummary)
		timeTracking.GET("/team-status", middleware.RequirePermission(roleRepo, "users.manage"), timeTrackingHandler.GetTeamStatus)
		timeTracking.GET("/team-overview", middleware.RequirePermission(roleRepo, "users.manage"), timeTrackingHandler.GetTeamOverview)
	}
}
