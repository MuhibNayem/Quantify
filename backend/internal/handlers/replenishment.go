package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
)

type ReplenishmentHandler struct {
	forecastingService   services.ForecastingService
	replenishmentService services.ReplenishmentService
}

func NewReplenishmentHandler(forecastingService services.ForecastingService, replenishmentService services.ReplenishmentService) *ReplenishmentHandler {
	return &ReplenishmentHandler{
		forecastingService:   forecastingService,
		replenishmentService: replenishmentService,
	}
}

// Mock storage for forecast jobs and POs
var forecastJobs = make(map[string]gin.H)
var purchaseOrders = make(map[uint]domain.PurchaseOrder) // Using PO ID as key

// GenerateDemandForecast godoc
// @Summary Trigger demand forecast generation
// @Description Triggers a demand forecasting process for a product or all products
// @Tags replenishment
// @Accept json
// @Produce json
// @Param forecast body requests.ForecastGenerationRequest true "Forecast generation request"
// @Success 200 {object} map[string]interface{} "Forecast generation status"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/forecast/generate [post]
func (h *ReplenishmentHandler) GenerateDemandForecast(c *gin.Context) {
	var req requests.ForecastGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if err := h.forecastingService.GenerateDemandForecast(req.ProductID, req.PeriodInDays); err != nil {
		c.Error(appErrors.NewAppError("Failed to generate demand forecast", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Demand forecast generation initiated."})
}

// GetDemandForecast godoc
// @Summary Get a specific demand forecast
// @Description Retrieves details of a specific demand forecast by its ID
// @Tags replenishment
// @Accept json
// @Produce json
// @Param forecastId path int true "Forecast ID"
// @Success 200 {object} domain.DemandForecast
// @Failure 404 {object} map[string]interface{} "Forecast not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/forecast/{forecastId} [get]
func GetDemandForecast(c *gin.Context) {
	forecastID := c.Param("forecastId") // This is a mock job ID, not DB ID
	// In a real scenario, you'd fetch from DB by actual forecast ID
	var forecast domain.DemandForecast
	if err := repository.DB.First(&forecast, forecastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Forecast not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch forecast", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, forecast)
}

// GenerateReorderSuggestions godoc
// @Summary Generate reorder suggestions
// @Description Triggers the generation of reorder suggestions based on current stock levels
// @Tags replenishment
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Generation status"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/suggestions/generate [post]
func (h *ReplenishmentHandler) GenerateReorderSuggestions(c *gin.Context) {
	if err := h.replenishmentService.GenerateReorderSuggestions(); err != nil {
		c.Error(appErrors.NewAppError("Failed to generate reorder suggestions", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reorder suggestions generated successfully"})
}

// ListReorderSuggestions godoc
// @Summary Get a list of reorder suggestions
// @Description Retrieves a list of suggested reorders based on forecast and stock levels
// @Tags replenishment
// @Accept json
// @Produce json
// @Param status query string false "Filter by suggestion status (PENDING, APPROVED)"
// @Param supplierId query int false "Filter by Supplier ID"
// @Success 200 {array} domain.ReorderSuggestion
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/suggestions [get]
func (h *ReplenishmentHandler) ListReorderSuggestions(c *gin.Context) {
	var suggestions []domain.ReorderSuggestion
	db := repository.DB.Preload("Product").Preload("Supplier")

	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if supplierID := c.Query("supplierId"); supplierID != "" {
		db = db.Where("supplier_id = ?", supplierID)
	}

	if err := db.Find(&suggestions).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch reorder suggestions", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, suggestions)
}

// CreatePOFromSuggestion godoc
// @Summary Create a draft Purchase Order from a reorder suggestion
// @Description Creates a draft Purchase Order based on a selected reorder suggestion
// @Tags replenishment
// @Accept json
// @Produce json
// @Param suggestionId path int true "Reorder Suggestion ID"
// @Success 201 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Suggestion not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/suggestions/{suggestionId}/create-po [post]
func CreatePOFromSuggestion(c *gin.Context) {
	suggestionID := c.Param("suggestionId")
	var suggestion domain.ReorderSuggestion
	if err := repository.DB.Preload("Product").Preload("Supplier").First(&suggestion, suggestionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Reorder suggestion not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch suggestion", http.StatusInternalServerError, err))
		return
	}

	if suggestion.Status != "PENDING" {
		c.Error(appErrors.NewAppError("Suggestion is not in PENDING state", http.StatusBadRequest, nil))
		return
	}

	po := domain.PurchaseOrder{
		SupplierID: suggestion.SupplierID,
		Status:     "DRAFT",
		OrderDate:  time.Now(),
		CreatedBy:  1, // TODO: Replace with actual UserID
		PurchaseOrderItems: []domain.PurchaseOrderItem{
			{
				ProductID:       suggestion.ProductID,
				OrderedQuantity: suggestion.SuggestedOrderQuantity,
				UnitPrice:       suggestion.Product.PurchasePrice,
			},
		},
	}

	if err := repository.DB.Create(&po).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err))
		return
	}

	// Update suggestion status
	if err := repository.DB.Model(&suggestion).Update("Status", "PO_CREATED").Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to update suggestion status", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, po)
}

// SendPurchaseOrder godoc
// @Summary Send a Purchase Order to the supplier
// @Description Marks an approved Purchase Order as SENT
// @Tags replenishment
// @Accept json
// @Produce json
// @Param poId path int true "Purchase Order ID"
// @Success 200 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Purchase Order not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /purchase-orders/{poId}/send [post]
func SendPurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var po domain.PurchaseOrder
	if err := repository.DB.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		return
	}

	if po.Status != "APPROVED" {
		c.Error(appErrors.NewAppError("Purchase Order is not in APPROVED state", http.StatusBadRequest, nil))
		return
	}

	if err := repository.DB.Model(&po).Update("Status", "SENT").Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to mark purchase order as sent", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, po)
}

// ApprovePurchaseOrder godoc
// @Summary Approve a draft Purchase Order
// @Description Approves a draft Purchase Order, changing its status to APPROVED
// @Tags replenishment
// @Accept json
// @Produce json
// @Param poId path int true "Purchase Order ID"
// @Success 200 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Purchase Order not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /purchase-orders/{poId}/approve [post]
func ApprovePurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var po domain.PurchaseOrder
	if err := repository.DB.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		return
	}

	if po.Status != "DRAFT" {
		c.Error(appErrors.NewAppError("Purchase Order is not in DRAFT state", http.StatusBadRequest, nil))
		return
	}

	if err := repository.DB.Model(&po).Updates(map[string]interface{}{
		"Status":     "APPROVED",
		"ApprovedBy": 1, // TODO: Replace with actual UserID
		"ApprovedAt": time.Now(),
	}).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to approve purchase order", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, po)
}

// GetPurchaseOrder godoc
// @Summary Get a Purchase Order by ID
// @Description Retrieves details of a specific Purchase Order by its ID
// @Tags replenishment
// @Accept json
// @Produce json
// @Param poId path int true "Purchase Order ID"
// @Success 200 {object} domain.PurchaseOrder
// @Failure 404 {object} map[string]interface{} "Purchase Order not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /purchase-orders/{poId} [get]
func GetPurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var po domain.PurchaseOrder
	if err := repository.DB.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, po)
}

// UpdatePurchaseOrder godoc
// @Summary Update a Purchase Order
// @Description Updates details of a specific Purchase Order
// @Tags replenishment
// @Accept json
// @Produce json
// @Param poId path int true "Purchase Order ID"
// @Param po body requests.UpdatePORequest true "Purchase Order update request"
// @Success 200 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Purchase Order not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /purchase-orders/{poId} [put]
func UpdatePurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var req requests.UpdatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var po domain.PurchaseOrder
	if err := repository.DB.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		return
	}

	// Only allow updates if PO is in DRAFT status
	if po.Status != "DRAFT" {
		c.Error(appErrors.NewAppError(fmt.Sprintf("Cannot update Purchase Order in %s status", po.Status), http.StatusConflict, nil))
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.SupplierID != 0 {
		updates["SupplierID"] = req.SupplierID
	}
	if req.Status != "" {
		updates["Status"] = req.Status
	}
	if req.OrderDate != nil {
		updates["OrderDate"] = *req.OrderDate
	}
	if req.ExpectedDeliveryDate != nil {
		updates["ExpectedDeliveryDate"] = *req.ExpectedDeliveryDate
	}

	if len(updates) > 0 {
		if err := repository.DB.Model(&po).Updates(updates).Error; err != nil {
			c.Error(appErrors.NewAppError("Failed to update purchase order", http.StatusInternalServerError, err))
			return
		}
	}

	// Update items if provided
	if len(req.PurchaseOrderItems) > 0 {
		// Delete existing items and create new ones for simplicity
		if err := repository.DB.Where("purchase_order_id = ?", po.ID).Delete(&domain.PurchaseOrderItem{}).Error; err != nil {
			c.Error(appErrors.NewAppError("Failed to clear existing PO items", http.StatusInternalServerError, err))
			return
		}
		for _, itemReq := range req.PurchaseOrderItems {
			poItem := domain.PurchaseOrderItem{
				PurchaseOrderID: po.ID,
				ProductID:       itemReq.ProductID,
				OrderedQuantity: itemReq.OrderedQuantity,
				UnitPrice:       itemReq.UnitPrice,
			}
			if err := repository.DB.Create(&poItem).Error; err != nil {
				c.Error(appErrors.NewAppError("Failed to create PO item", http.StatusInternalServerError, err))
				return
			}
		}
	}

	// Reload PO with updated items
	repository.DB.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID)
	c.JSON(http.StatusOK, po)
}

// ReceivePurchaseOrder godoc
// @Summary Record received goods for a Purchase Order
// @Description Records received quantities for items in a Purchase Order and updates stock levels
// @Tags replenishment
// @Accept json
// @Produce json
// @Param poId path int true "Purchase Order ID"
// @Param receivedItems body requests.ReceivePORequest true "Received items details"
// @Success 200 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Purchase Order not found"
// @Failure 409 {object} map[string]interface{} "Conflict: PO not in SENT status or items already fully received"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /purchase-orders/{poId}/receive [post]
func ReceivePurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var req requests.ReceivePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var po domain.PurchaseOrder
	if err := repository.DB.Preload("PurchaseOrderItems").First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		return
	}

	// Only allow receiving if PO is in APPROVED or SENT status
	if po.Status != "APPROVED" && po.Status != "SENT" {
		c.Error(appErrors.NewAppError(fmt.Sprintf("Cannot receive goods for Purchase Order in %s status", po.Status), http.StatusConflict, nil))
		return
	}

	// Use a transaction for receiving goods to ensure atomicity
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		for _, receivedItem := range req.Items {
			var poItem domain.PurchaseOrderItem
			if err := tx.First(&poItem, receivedItem.PurchaseOrderItemID).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Purchase Order Item %d not found", receivedItem.PurchaseOrderItemID), http.StatusNotFound, err)
			}

			if poItem.PurchaseOrderID != po.ID {
				return appErrors.NewAppError("Purchase Order Item does not belong to this Purchase Order", http.StatusBadRequest, nil)
			}

			// Check if received quantity exceeds ordered quantity
			if receivedItem.ReceivedQuantity > (poItem.OrderedQuantity - poItem.ReceivedQuantity) {
				return appErrors.NewAppError(fmt.Sprintf("Received quantity %d for item %d exceeds remaining ordered quantity %d", receivedItem.ReceivedQuantity, poItem.ID, (poItem.OrderedQuantity-poItem.ReceivedQuantity)), http.StatusBadRequest, nil)
			}

			// Update received quantity for PO item
			poItem.ReceivedQuantity += receivedItem.ReceivedQuantity
			if err := tx.Save(&poItem).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to update received quantity for PO item %d", poItem.ID), http.StatusInternalServerError, err)
			}

			// Create new batch and update stock
			batch := domain.Batch{
				ProductID:   poItem.ProductID,
				BatchNumber: receivedItem.BatchNumber,
				Quantity:    receivedItem.ReceivedQuantity,
				ExpiryDate:  receivedItem.ExpiryDate,
			}
			if err := tx.Create(&batch).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to create batch for product %d", poItem.ProductID), http.StatusInternalServerError, err)
			}
		}

		// Check if all items are fully received to update PO status to RECEIVED
		allReceived := true
		for _, item := range po.PurchaseOrderItems {
			if item.OrderedQuantity != item.ReceivedQuantity {
				allReceived = false
				break
			}
		}

		if allReceived {
			if err := tx.Model(&po).Updates(map[string]interface{}{"Status": "RECEIVED", "ActualDeliveryDate": time.Now()}).Error; err != nil {
				return appErrors.NewAppError("Failed to update PO status to RECEIVED", http.StatusInternalServerError, err)
			}
		} else {
			if err := tx.Model(&po).Update("Status", "PARTIALLY_RECEIVED").Error; err != nil {
				return appErrors.NewAppError("Failed to update PO status to PARTIALLY_RECEIVED", http.StatusInternalServerError, err)
			}
		}
		return nil
	})

	if err != nil {
		// If the error is an AppError, it's already formatted
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to receive purchase order goods", http.StatusInternalServerError, err))
		}
		return
	}

	// Reload PO with updated items
	repository.DB.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID)
	c.JSON(http.StatusOK, po)
}

// CancelPurchaseOrder godoc
// @Summary Cancel a Purchase Order
// @Description Cancels a Purchase Order if it's in DRAFT or APPROVED status
// @Tags replenishment
// @Accept json
// @Produce json
// @Param poId path int true "Purchase Order ID"
// @Success 200 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Purchase Order not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /purchase-orders/{poId}/cancel [post]
func CancelPurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var po domain.PurchaseOrder
	if err := repository.DB.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		return
	}

	// Only allow cancellation if PO is in DRAFT or APPROVED status
	if po.Status != "DRAFT" && po.Status != "APPROVED" {
		c.Error(appErrors.NewAppError(fmt.Sprintf("Cannot cancel Purchase Order in %s status", po.Status), http.StatusConflict, nil))
		return
	}

	if err := repository.DB.Model(&po).Update("Status", "CANCELLED").Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to cancel purchase order", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, po)
}
