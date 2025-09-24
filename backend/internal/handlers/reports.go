package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

// GetSalesTrendsReport godoc
// @Summary Get sales trends report
// @Description Generates a report on sales trends over a specified period, with optional filters.
// @Tags reports
// @Accept json
// @Produce json
// @Param request body requests.SalesTrendsReportRequest true "Sales trends report parameters"
// @Success 200 {object} map[string]interface{} "Sales trends data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/sales-trends [post]
func GetSalesTrendsReport(c *gin.Context) {
	var req requests.SalesTrendsReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	logrus.Infof("Generating sales trends report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	salesTrends, topSellingProducts, err := repository.GetSalesTrends(req.StartDate, req.EndDate, req.CategoryID, req.LocationID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get sales trends", http.StatusInternalServerError, err))
		return
	}

	var totalSales float64
	for _, trend := range salesTrends {
		totalSales += trend.TotalSales
	}

	days := req.EndDate.Sub(req.StartDate).Hours() / 24
	if days == 0 {
		days = 1
	}
	averageDailySales := totalSales / days

	reportData := gin.H{
		"period":             fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		"totalSales":         totalSales,
		"averageDailySales":  averageDailySales,
		"salesTrends":        salesTrends,
		"topSellingProducts": topSellingProducts,
	}

	c.JSON(http.StatusOK, reportData)
}

// GetInventoryTurnoverReport godoc
// @Summary Get inventory turnover report
// @Description Generates a report on inventory turnover rate over a specified period.
// @Tags reports
// @Accept json
// @Produce json
// @Param request body requests.InventoryTurnoverReportRequest true "Inventory turnover report parameters"
// @Success 200 {object} map[string]interface{} "Inventory turnover data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/inventory-turnover [post]
func GetInventoryTurnoverReport(c *gin.Context) {
	var req requests.InventoryTurnoverReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	logrus.Infof("Generating inventory turnover report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	costOfGoodsSold, averageInventoryValue, err := repository.GetInventoryTurnover(req.StartDate, req.EndDate, req.CategoryID, req.LocationID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get inventory turnover", http.StatusInternalServerError, err))
		return
	}

	var inventoryTurnoverRate float64
	if averageInventoryValue > 0 {
		inventoryTurnoverRate = costOfGoodsSold / averageInventoryValue
	}

	reportData := gin.H{
		"period":                  fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		"totalCostOfGoodsSold":    costOfGoodsSold,
		"averageInventoryValue":   averageInventoryValue,
		"inventoryTurnoverRate":   inventoryTurnoverRate,
	}

	c.JSON(http.StatusOK, reportData)
}

// GetProfitMarginReport godoc
// @Summary Get profit margin report
// @Description Generates a report on profit margins for products or categories over a specified period.
// @Tags reports
// @Accept json
// @Produce json
// @Param request body requests.ProfitMarginReportRequest true "Profit margin report parameters"
// @Success 200 {object} map[string]interface{} "Profit margin data"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reports/profit-margin [post]
func GetProfitMarginReport(c *gin.Context) {
	var req requests.ProfitMarginReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	logrus.Infof("Generating profit margin report for period %v to %v, category %v, location %v", req.StartDate, req.EndDate, req.CategoryID, req.LocationID)

	totalRevenue, totalCost, err := repository.GetProfitMargin(req.StartDate, req.EndDate, req.CategoryID, req.LocationID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get profit margin", http.StatusInternalServerError, err))
		return
	}

	grossProfit := totalRevenue - totalCost
	var grossProfitMargin float64
	if totalRevenue > 0 {
		grossProfitMargin = grossProfit / totalRevenue
	}

	reportData := gin.H{
		"period":            fmt.Sprintf("%s to %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		"totalRevenue":      totalRevenue,
		"totalCost":         totalCost,
		"grossProfit":       grossProfit,
		"grossProfitMargin": grossProfitMargin,
	}

	c.JSON(http.StatusOK, reportData)
}

func GenerateDailySalesSummary() {
	yesterday := time.Now().AddDate(0, 0, -1)
	totalSales, err := repository.GetDailySalesSummary(yesterday)
	if err != nil {
		logrus.Errorf("Failed to generate daily sales summary: %v", err)
		return
	}

	logrus.Infof("Daily sales summary generated for %s: Total Sales = $%.2f", yesterday.Format("2006-01-02"), totalSales)
}
