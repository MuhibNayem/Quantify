package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterFinance(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	financeHandler := handlers.NewFinanceHandler(c.DB)

	finance := rg.Group("/finance")
	{
		finance.GET("/trial-balance", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetTrialBalance)
		finance.GET("/income-statement", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetIncomeStatement)
		finance.GET("/balance-sheet", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetBalanceSheet)
		finance.GET("/ar-aging", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetARAging)
		finance.GET("/receivables", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetOutstandingAR)
		finance.POST("/payments", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.RecordPayment)
		finance.GET("/ap-aging", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetAPAging)
		finance.GET("/payables", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.GetOutstandingAP)
		finance.POST("/supplier-payments", middleware.RequirePermission(roleRepo, "reports.financial"), financeHandler.RecordSupplierPayment)
	}
}
