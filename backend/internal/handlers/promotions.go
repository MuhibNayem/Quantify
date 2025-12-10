package handlers

import (
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PromotionHandler struct {
	DB *gorm.DB
}

func NewPromotionHandler(db *gorm.DB) *PromotionHandler {
	return &PromotionHandler{DB: db}
}

// CreatePromotion godoc
// @Summary Create a new promotion
// @Description Create a new discount rule
// @Tags promotions
// @Accept json
// @Produce json
// @Param promotion body requests.PromotionCreateRequest true "Promotion creation request"
// @Success 201 {object} domain.Promotion
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /promotions [post]
func (h *PromotionHandler) CreatePromotion(c *gin.Context) {
	var req requests.PromotionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	promotion := domain.Promotion{
		Name:          req.Name,
		Description:   req.Description,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Priority:      req.Priority,
		ProductID:     req.ProductID,
		CategoryID:    req.CategoryID,
		SubCategoryID: req.SubCategoryID,
		IsActive:      true,
	}

	if err := h.DB.Create(&promotion).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to create promotion", http.StatusInternalServerError, err))
		return
	}

	repository.DeleteCache("promotions:*")

	c.JSON(http.StatusCreated, promotion)
}

// ListPromotions godoc
// @Summary List promotions
// @Description Get a list of all promotions
// @Tags promotions
// @Accept json
// @Produce json
// @Success 200 {array} domain.Promotion
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /promotions [get]
func (h *PromotionHandler) ListPromotions(c *gin.Context) {
	var promotions []domain.Promotion

	query := h.DB.Preload("Product").Preload("Category").Preload("SubCategory").Order("priority desc, created_at desc")

	if active := c.Query("active"); active == "true" {
		now := time.Now()
		query = query.Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, now, now)
	}

	if err := query.Find(&promotions).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch promotions", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"promotions": promotions})
}

// UpdatePromotion godoc
// @Summary Update a promotion
// @Description Update an existing promotion
// @Tags promotions
// @Accept json
// @Produce json
// @Param id path int true "Promotion ID"
// @Param promotion body requests.PromotionUpdateRequest true "Promotion update request"
// @Success 200 {object} domain.Promotion
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /promotions/{id} [put]
func (h *PromotionHandler) UpdatePromotion(c *gin.Context) {
	id := c.Param("id")
	var req requests.PromotionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var promotion domain.Promotion
	if err := h.DB.First(&promotion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Promotion not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch promotion", http.StatusInternalServerError, err))
		return
	}

	updates := map[string]interface{}{
		"Priority": req.Priority,
	}
	if req.Name != "" {
		updates["Name"] = req.Name
	}
	if req.Description != "" {
		updates["Description"] = req.Description
	}
	if req.DiscountType != "" {
		updates["DiscountType"] = req.DiscountType
	}
	if req.DiscountValue > 0 {
		updates["DiscountValue"] = req.DiscountValue
	}
	if !req.StartDate.IsZero() {
		updates["StartDate"] = req.StartDate
	}
	if !req.EndDate.IsZero() {
		updates["EndDate"] = req.EndDate
	}
	if req.IsActive != nil {
		updates["IsActive"] = *req.IsActive
	}

	if err := h.DB.Model(&promotion).Updates(updates).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to update promotion", http.StatusInternalServerError, err))
		return
	}

	repository.DeleteCache("promotions:*")

	c.JSON(http.StatusOK, promotion)
}

// DeletePromotion godoc
// @Summary Delete a promotion
// @Description Soft delete a promotion
// @Tags promotions
// @Accept json
// @Produce json
// @Param id path int true "Promotion ID"
// @Success 204 "No Content"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /promotions/{id} [delete]
func (h *PromotionHandler) DeletePromotion(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&domain.Promotion{}, id).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete promotion", http.StatusInternalServerError, err))
		return
	}

	repository.DeleteCache("promotions:*")

	c.Status(http.StatusNoContent)
}
