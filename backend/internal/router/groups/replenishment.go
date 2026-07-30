package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterReplenishment(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	replenishmentHandler := c.ReplenishmentHandler

	replenishment := rg.Group("/replenishment")
	{
		replenishment.POST("/forecast/generate", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.GenerateDemandForecast)
		replenishment.GET("/forecast/:forecastId", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.GetDemandForecast)
		replenishment.GET("/dashboard", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.GetForecastDashboard)
		replenishment.POST("/suggestions/generate", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.GenerateReorderSuggestions)
		replenishment.GET("/suggestions", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.ListReorderSuggestions)
		replenishment.POST("/suggestions/:suggestionId/create-po", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.CreatePOFromSuggestion)
		replenishment.POST("/purchase-orders/:poId/approve", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.ApprovePurchaseOrder)
		replenishment.POST("/purchase-orders/:poId/send", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.SendPurchaseOrder)
		replenishment.GET("/purchase-orders/:poId", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.GetPurchaseOrder)
		replenishment.PUT("/purchase-orders/:poId", middleware.RequirePermission(roleRepo, "inventory.write"), replenishmentHandler.UpdatePurchaseOrder)
		replenishment.POST("/purchase-orders/:poId/receive", middleware.RequirePermission(roleRepo, "inventory.write"), replenishmentHandler.ReceivePurchaseOrder)
		replenishment.POST("/purchase-orders/:poId/cancel", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.CancelPurchaseOrder)
		replenishment.POST("/purchase-orders", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.CreatePurchaseOrder)
		replenishment.GET("/purchase-orders", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.ListPurchaseOrders)
		replenishment.POST("/returns", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.CreatePurchaseReturn)
		replenishment.GET("/returns", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.ListPurchaseReturns)
		replenishment.GET("/returns/:id", middleware.RequirePermission(roleRepo, "replenishment.read"), replenishmentHandler.GetPurchaseReturn)
		replenishment.POST("/returns/:id/approve", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.ApprovePurchaseReturn)
		replenishment.POST("/returns/:id/complete", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.CompletePurchaseReturn)
		replenishment.POST("/returns/:id/cancel", middleware.RequirePermission(roleRepo, "replenishment.write"), replenishmentHandler.CancelPurchaseReturn)
	}
}
