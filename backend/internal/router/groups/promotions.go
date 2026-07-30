package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterPromotions(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	promotionHandler := c.Promotion

	promotions := rg.Group("/promotions")
	{
		promotions.POST("", middleware.RequirePermission(roleRepo, "products.write"), promotionHandler.CreatePromotion)
		promotions.GET("", middleware.RequirePermission(roleRepo, "products.read"), promotionHandler.ListPromotions)
		promotions.PUT("/:id", middleware.RequirePermission(roleRepo, "products.write"), promotionHandler.UpdatePromotion)
		promotions.DELETE("/:id", middleware.RequirePermission(roleRepo, "products.delete"), promotionHandler.DeletePromotion)
	}
}
