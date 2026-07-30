package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterReports(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	reportHandler := c.Report

	reports := rg.Group("/reports")
	{
		reports.POST("/sales-trends", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetSalesTrendsReport)
		reports.POST("/sales-trends/export", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.ExportSalesTrendsReport)
		reports.GET("/jobs/:jobId", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetReportJobStatus)
		reports.GET("/download/:jobId", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.DownloadReportFile)
		reports.POST("/inventory-turnover", middleware.RequirePermission(roleRepo, "reports.inventory"), reportHandler.GetInventoryTurnoverReport)
		reports.POST("/profit-margin", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetProfitMarginReport)

		reports.GET("/stock-aging", middleware.RequirePermission(roleRepo, "reports.inventory"), reportHandler.GetStockAgingReport)
		reports.GET("/dead-stock", middleware.RequirePermission(roleRepo, "reports.inventory"), reportHandler.GetDeadStockReport)
		reports.GET("/supplier-performance", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetSupplierPerformanceReport)
		reports.GET("/heatmap", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetHourlySalesHeatmap)
		reports.GET("/employee-sales", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetSalesByEmployeeReport)
		reports.GET("/category-drilldown", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetCategoryDrillDownReport)
		reports.GET("/gmroi", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetCOGSAndGMROIReport)
		reports.GET("/audit/voids", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetVoidDiscountAuditReport)
		reports.GET("/tax-liability", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetTaxLiabilityReport)
		reports.GET("/cash-reconciliation", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.GetCashDrawerReconciliationReport)

		reports.GET("/customer-insights", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetCustomerInsightsReport)
		reports.GET("/shrinkage", middleware.RequirePermission(roleRepo, "reports.inventory"), reportHandler.GetShrinkageReport)
		reports.GET("/returns-analysis", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetCustomerReturnAnalysisReport)
		reports.GET("/basket-analysis", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetBasketAnalysisReport)
		reports.GET("/product-performance", middleware.RequirePermission(roleRepo, "reports.sales"), reportHandler.GetProductPerformanceAnalytics)
	}

	rg.GET("/audit-logs", middleware.RequirePermission(roleRepo, "reports.financial"), reportHandler.ListAuditLogs)
}
