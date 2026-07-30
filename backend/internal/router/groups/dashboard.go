package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterDashboard(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	dashboardHandler := c.DashboardHandler

	rg.GET("/dashboard/summary", middleware.RequirePermission(roleRepo, "dashboard.view"), dashboardHandler.GetDashboardSummary)
}
