package groups

import (
	"inventory/backend/internal/app"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterCategories(rg *gin.RouterGroup, c *app.Container, roleRepo repository.RoleRepository) {
	categoryHandler := c.Category

	categories := rg.Group("/categories")
	{
		categories.POST("", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.CreateCategory)
		categories.GET("", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.ListCategories)
		categories.GET("/:categoryId", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.GetCategory)
		categories.GET("/name/:name", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.GetCategoryByName)
		categories.PUT("/:categoryId", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.UpdateCategory)
		categories.DELETE("/:categoryId", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.DeleteCategory)
		categories.POST("/:categoryId/sub-categories", middleware.RequirePermission(roleRepo, "categories.write"), categoryHandler.CreateSubCategory)
		categories.GET("/:categoryId/sub-categories", middleware.RequirePermission(roleRepo, "categories.read"), categoryHandler.ListSubCategories)
	}
}
