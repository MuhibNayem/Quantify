package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterTimeTracking(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	timeTrackingHandler := c.TimeTrackingHandler

	selfOrManager := middleware.SelfOrPermission(roleRepo, "users.manage")
	managerOnly := middleware.RequirePermission(roleRepo, "users.manage")

	timeTracking := rg.Group("/time-tracking")
	{
		timeTracking.POST("/clock-in/:userId", selfOrManager, timeTrackingHandler.ClockIn)
		timeTracking.POST("/clock-out/:userId", selfOrManager, timeTrackingHandler.ClockOut)
		timeTracking.POST("/break-start/:userId", selfOrManager, timeTrackingHandler.StartBreak)
		timeTracking.POST("/break-end/:userId", selfOrManager, timeTrackingHandler.EndBreak)
		timeTracking.GET("/last-entry/:userId", selfOrManager, timeTrackingHandler.GetLastTimeClock)
		timeTracking.GET("/history/:userId", selfOrManager, timeTrackingHandler.GetHistory)
		timeTracking.GET("/activities", managerOnly, timeTrackingHandler.GetRecentActivities)
		timeTracking.GET("/weekly-summary/:userId", selfOrManager, timeTrackingHandler.GetWeeklySummary)
		timeTracking.GET("/team-status", managerOnly, timeTrackingHandler.GetTeamStatus)
		timeTracking.GET("/team-overview", managerOnly, timeTrackingHandler.GetTeamOverview)
	}
}
