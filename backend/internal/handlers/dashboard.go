package handlers

import (
	"net/http"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	repo *repository.DashboardRepository
}

func NewDashboardHandler(repo *repository.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

// GetDashboardSummary godoc
// @Summary Get dashboard summary data
// @Description Retrieves aggregated statistics and recent items for the dashboard
// @Tags dashboard
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /dashboard/summary [get]
func (h *DashboardHandler) GetDashboardSummary(c *gin.Context) {
	stats, err := h.repo.GetStats()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch dashboard stats", http.StatusInternalServerError, err))
		return
	}

	recentProducts, err := h.repo.GetRecentProducts(5)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch recent products", http.StatusInternalServerError, err))
		return
	}

	recentAlerts, err := h.repo.GetRecentAlerts(5)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch recent alerts", http.StatusInternalServerError, err))
		return
	}

	suggestions, err := h.repo.GetRecentReorderSuggestions(5)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch reorder suggestions", http.StatusInternalServerError, err))
		return
	}

	chartData, err := h.repo.GetSalesChartData()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch sales chart data", http.StatusInternalServerError, err))
		return
	}

	trend, err := h.repo.GetSalesTrend()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch sales trend", http.StatusInternalServerError, err))
		return
	}

	// Transform suggestions to match frontend expectation if needed,
	// but sending domain objects is usually fine if json tags match.

	response := gin.H{
		"stats": gin.H{
			"products":   stats.ProductCount,
			"categories": stats.CategoryCount,
			"suppliers":  stats.SupplierCount,
			"alerts":     stats.ActiveAlertCount,
		},
		"recentProducts": recentProducts,
		"recentAlerts":   recentAlerts,
		"suggestions":    suggestions,
		"chartData":      chartData,
		"trend":          trend,
	}

	c.JSON(http.StatusOK, response)
}
