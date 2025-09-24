package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new product category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body requests.CategoryCreateRequest true "Category creation request"
// @Success 201 {object} domain.Category
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 409 {object} map[string]interface{} "Conflict: Category with this name already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories [post]
func CreateCategory(c *gin.Context) {
	var req requests.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	category := domain.Category{
		Name: req.Name,
	}

	if err := repository.DB.Create(&category).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Category with this name already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to create category", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, category)
}

// ListCategories godoc
// @Summary Get a list of categories
// @Description Get a list of all product categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} domain.Category
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories [get]
func ListCategories(c *gin.Context) {
	var categories []domain.Category
	if err := repository.DB.Preload("SubCategories").Find(&categories).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch categories", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetCategory godoc
// @Summary Get a category by ID
// @Description Get a single category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} domain.Category
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories/{id} [get]
func GetCategory(c *gin.Context) {
	id := c.Param("id")
	var category domain.Category
	if err := repository.DB.Preload("SubCategories").First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch category", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, category)
}

// UpdateCategory godoc
// @Summary Update an existing category
// @Description Update an existing product category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body requests.CategoryUpdateRequest true "Category update request"
// @Success 200 {object} domain.Category
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var req requests.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var category domain.Category
	if err := repository.DB.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch category", http.StatusInternalServerError, err))
		return
	}

	if err := repository.DB.Model(&category).Update("Name", req.Name).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Category with this name already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to update category", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a product category by its ID. Cannot delete if products or subcategories are associated.
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Category has associated data"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category domain.Category
	if err := repository.DB.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch category", http.StatusInternalServerError, err))
		return
	}

	// Check for associated products
	var productCount int64
	repository.DB.Model(&domain.Product{}).Where("category_id = ?", id).Count(&productCount)
	if productCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete category: products are associated", http.StatusConflict, nil))
		return
	}

	// Check for associated subcategories
	var subCategoryCount int64
	repository.DB.Model(&domain.SubCategory{}).Where("category_id = ?", id).Count(&subCategoryCount)
	if subCategoryCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete category: subcategories are associated", http.StatusConflict, nil))
		return
	}

	if err := repository.DB.Delete(&category).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete category", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateSubCategory godoc
// @Summary Create a new sub-category
// @Description Create a new sub-category for a specific category
// @Tags categories
// @Accept json
// @Produce json
// @Param categoryId path int true "Category ID"
// @Param subCategory body requests.SubCategoryCreateRequest true "Sub-category creation request"
// @Success 201 {object} domain.SubCategory
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories/{categoryId}/subcategories [post]
func CreateSubCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("categoryId"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid category ID", http.StatusBadRequest, err))
		return
	}

	var req requests.SubCategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var category domain.Category
	if err := repository.DB.First(&category, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch category", http.StatusInternalServerError, err))
		return
	}

	subCategory := domain.SubCategory{
		Name:       req.Name,
		CategoryID: uint(categoryID),
	}

	if err := repository.DB.Create(&subCategory).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to create sub-category", http.StatusInternalServerError, err))
		return	}

	c.JSON(http.StatusCreated, subCategory)
}

// ListSubCategories godoc
// @Summary Get sub-categories for a category
// @Description Get a list of sub-categories for a specific category
// @Tags categories
// @Accept json
// @Produce json
// @Param categoryId path int true "Category ID"
// @Success 200 {array} domain.SubCategory
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /categories/{categoryId}/subcategories [get]
func ListSubCategories(c *gin.Context) {
	categoryID := c.Param("categoryId")
	var subCategories []domain.SubCategory
	if err := repository.DB.Where("category_id = ?", categoryID).Find(&subCategories).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch sub-categories", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, subCategories)
}

// GetSubCategory godoc
// @Summary Get a sub-category by ID
// @Description Get a single sub-category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Sub-Category ID"
// @Success 200 {object} domain.SubCategory
// @Failure 404 {object} map[string]interface{} "Sub-Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /subcategories/{id} [get]
func GetSubCategory(c *gin.Context) {
	id := c.Param("id")
	var subCategory domain.SubCategory
	if err := repository.DB.First(&subCategory, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Sub-Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch sub-category", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, subCategory)
}

// UpdateSubCategory godoc
// @Summary Update an existing sub-category
// @Description Update an existing sub-category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Sub-Category ID"
// @Param subCategory body requests.SubCategoryUpdateRequest true "Sub-category update request"
// @Success 200 {object} domain.SubCategory
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Sub-Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /subcategories/{id} [put]
func UpdateSubCategory(c *gin.Context) {
	id := c.Param("id")
	var req requests.SubCategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var subCategory domain.SubCategory
	if err := repository.DB.First(&subCategory, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Sub-Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch sub-category", http.StatusInternalServerError, err))
		return
	}

	if err := repository.DB.Model(&subCategory).Update("Name", req.Name).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to update sub-category", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, subCategory)
}

// DeleteSubCategory godoc
// @Summary Delete a sub-category
// @Description Delete a sub-category by its ID. Cannot delete if products are associated.
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Sub-Category ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{} "Sub-Category not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Sub-Category has associated products"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /subcategories/{id} [delete]
func DeleteSubCategory(c *gin.Context) {
	id := c.Param("id")
	var subCategory domain.SubCategory
	if err := repository.DB.First(&subCategory, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Sub-Category not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch sub-category", http.StatusInternalServerError, err))
		return
	}

	// Check for associated products
	var productCount int64
	repository.DB.Model(&domain.Product{}).Where("sub_category_id = ?", id).Count(&productCount)
	if productCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete sub-category: products are associated", http.StatusConflict, nil))
		return
	}

	if err := repository.DB.Delete(&subCategory).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete sub-category", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}
