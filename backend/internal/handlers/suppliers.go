package handlers

import (
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SupplierHandler struct {
	supplierRepo *repository.SupplierRepository
	db           *gorm.DB // Keep db for now for existing functions
}

func NewSupplierHandler(supplierRepo *repository.SupplierRepository, db *gorm.DB) *SupplierHandler {
	return &SupplierHandler{
		supplierRepo: supplierRepo,
		db:           db,
	}
}

// CreateSupplier godoc
// @Summary Create a new supplier
// @Description Create a new product supplier
// @Tags suppliers
// @Accept json
// @Produce json
// @Param supplier body requests.SupplierCreateRequest true "Supplier creation request"
// @Success 201 {object} domain.Supplier
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 409 {object} map[string]interface{} "Conflict: Supplier with this name already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers [post]
func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	var req requests.SupplierCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	supplier := domain.Supplier{
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
	}

	if err := h.supplierRepo.CreateSupplier(&supplier); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Supplier with this name already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to create supplier", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, supplier)
}

// ListSuppliers godoc
// @Summary Get a list of suppliers
// @Description Get a list of all product suppliers
// @Tags suppliers
// @Accept json
// @Produce json
// @Success 200 {array} domain.Supplier
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers [get]
func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	var suppliers []domain.Supplier
	if err := h.db.Find(&suppliers).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch suppliers", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, suppliers)
}

// GetSupplier godoc
// @Summary Get a supplier by ID
// @Description Get a single supplier by its ID
// @Tags suppliers
// @Accept json
// @Produce json
// @Param id path int true "Supplier ID"
// @Success 200 {object} domain.Supplier
// @Failure 404 {object} map[string]interface{} "Supplier not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers/{id} [get]
func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	id := c.Param("id")
	var supplier domain.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Supplier not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch supplier", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, supplier)
}

// GetSupplierByName godoc
// @Summary Get a supplier by Name
// @Description Get a single supplier by its Name
// @Tags suppliers
// @Accept json
// @Produce json
// @Param name path string true "Supplier Name"
// @Success 200 {object} domain.Supplier
// @Failure 404 {object} map[string]interface{} "Supplier not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers/name/{name} [get]
func (h *SupplierHandler) GetSupplierByName(c *gin.Context) {
	name := c.Param("name")
	supplier, err := h.supplierRepo.GetSupplierByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Supplier not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch supplier", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, supplier)
}

// UpdateSupplier godoc
// @Summary Update an existing supplier
// @Description Update an existing product supplier
// @Tags suppliers
// @Accept json
// @Produce json
// @Param id path int true "Supplier ID"
// @Param supplier body requests.SupplierUpdateRequest true "Supplier update request"
// @Success 200 {object} domain.Supplier
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Supplier not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers/{id} [put]
func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	id := c.Param("id")
	var req requests.SupplierUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var supplier domain.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Supplier not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch supplier for update", http.StatusInternalServerError, err))
		return
	}

	updates := map[string]interface{}{
		"Name":          req.Name,
		"ContactPerson": req.ContactPerson,
		"Email":         req.Email,
		"Phone":         req.Phone,
		"Address":       req.Address,
	}

	if err := h.supplierRepo.UpdateSupplier(&supplier, updates); err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.Error(appErrors.NewAppError("Supplier with this name already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to update supplier", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, supplier)
}

// DeleteSupplier godoc
// @Summary Delete a supplier
// @Description Delete a product supplier by its ID. Cannot delete if products are associated.
// @Tags suppliers
// @Accept json
// @Produce json
// @Param id path int true "Supplier ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{} "Supplier not found"
// @Failure 409 {object} map[string]interface{} "Conflict: Supplier has associated products"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers/{id} [delete]
func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	id := c.Param("id")
	var supplier domain.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Supplier not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch supplier", http.StatusInternalServerError, err))
		return
	}

	// Check for associated products
	var productCount int64
	h.db.Model(&domain.Product{}).Where("supplier_id = ?", id).Count(&productCount)
	if productCount > 0 {
		c.Error(appErrors.NewAppError("Cannot delete supplier: products are associated", http.StatusConflict, nil))
		return
	}

	if err := h.supplierRepo.DeleteSupplier(&supplier); err != nil {
		c.Error(appErrors.NewAppError("Failed to delete supplier", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetSupplierPerformanceReport godoc
// @Summary Get supplier performance report
// @Description Generates a mock report on supplier performance (e.g., on-time delivery, quality).
// @Tags suppliers
// @Accept json
// @Produce json
// @Param id path int true "Supplier ID"
// @Success 200 {object} map[string]interface{} "Supplier performance data"
// @Failure 404 {object} map[string]interface{} "Supplier not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /suppliers/{id}/performance [get]
func (h *SupplierHandler) GetSupplierPerformanceReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid supplier ID", http.StatusBadRequest, err))
		return
	}

	var supplier domain.Supplier
	if err := h.db.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Supplier not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch supplier", http.StatusInternalServerError, err))
		return
	}

	averageLeadTime, onTimeDeliveryRate, err := h.supplierRepo.GetSupplierPerformance(uint(id))
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get supplier performance", http.StatusInternalServerError, err))
		return
	}

	performanceData := gin.H{
		"supplierId":         supplier.ID,
		"supplierName":       supplier.Name,
		"averageLeadTimeDays": averageLeadTime,
		"onTimeDeliveryRate":   onTimeDeliveryRate,
	}

	c.JSON(http.StatusOK, performanceData)
}
