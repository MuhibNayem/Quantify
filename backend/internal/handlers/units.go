package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

// UnitsHandler exposes the unit-of-measure catalog and per-product selling units.
type UnitsHandler struct {
	DB *gorm.DB
}

func NewUnitsHandler(db *gorm.DB) *UnitsHandler {
	return &UnitsHandler{DB: db}
}

// ListUnits returns all units with their dimension.
func (h *UnitsHandler) ListUnits(c *gin.Context) {
	var units []domain.Unit
	if err := h.DB.Preload("Dimension").Order("dimension_id, code").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, units)
}

// ListDimensions returns all unit dimensions.
func (h *UnitsHandler) ListDimensions(c *gin.Context) {
	var dims []domain.UnitDimension
	if err := h.DB.Order("code").Find(&dims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dims)
}

// ListProductUnits returns the selling units configured for a product.
func (h *UnitsHandler) ListProductUnits(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var pus []domain.ProductUnit
	if err := h.DB.Preload("Unit").Where("product_id = ?", productID).Find(&pus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pus)
}

// ProductUnitRequest is the body for creating/updating a product selling unit.
type ProductUnitRequest struct {
	UnitID        uint    `json:"unitId" binding:"required"`
	Factor        float64 `json:"factor" binding:"required,gt=0"`
	Price         float64 `json:"price"`
	PurchasePrice float64 `json:"purchasePrice"`
	Barcode       string  `json:"barcode"`
	IsActive      bool    `json:"isActive"`
}

// SaveProductUnit creates or updates a selling unit for a product (upsert by product+unit).
func (h *UnitsHandler) SaveProductUnit(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req ProductUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pu := domain.ProductUnit{
		ProductID:     uint(productID),
		UnitID:        req.UnitID,
		Factor:        req.Factor,
		Price:         req.Price,
		PurchasePrice: req.PurchasePrice,
		Barcode:       req.Barcode,
		IsActive:      req.IsActive,
	}
	// Upsert on (product_id, unit_id).
	result := h.DB.Where("product_id = ? AND unit_id = ?", productID, req.UnitID).First(&domain.ProductUnit{})
	if result.Error == gorm.ErrRecordNotFound {
		if err := h.DB.Create(&pu).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		if err := h.DB.Model(&domain.ProductUnit{}).
			Where("product_id = ? AND unit_id = ?", productID, req.UnitID).
			Updates(map[string]interface{}{
				"factor": req.Factor, "price": req.Price, "purchase_price": req.PurchasePrice,
				"barcode": req.Barcode, "is_active": req.IsActive,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, pu)
}

// DeleteProductUnit removes a selling unit from a product.
func (h *UnitsHandler) DeleteProductUnit(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("productId"), 10, 64)
	unitID, _ := strconv.ParseUint(c.Param("unitId"), 10, 64)
	if err := h.DB.Where("product_id = ? AND unit_id = ?", productID, unitID).
		Delete(&domain.ProductUnit{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
