package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterSubCategories(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	categoryHandler := c.Category

	subCategories := rg.Group("/sub-categories")
	{
		subCategories.PUT("/:id", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.UpdateSubCategory)
		subCategories.DELETE("/:id", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.DeleteSubCategory)
	}
}
