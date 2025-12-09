package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
	"inventory/backend/internal/websocket"
)

type ReplenishmentHandler struct {
	ForecastingService   services.ForecastingService
	ReplenishmentService services.ReplenishmentService
	Hub                  *websocket.Hub
	NotificationRepo     repository.NotificationRepository
}

func NewReplenishmentHandler(forecastingService services.ForecastingService, replenishmentService services.ReplenishmentService, hub *websocket.Hub, notificationRepo repository.NotificationRepository) *ReplenishmentHandler {
	return &ReplenishmentHandler{
		ForecastingService:   forecastingService,
		ReplenishmentService: replenishmentService,
		Hub:                  hub,
		NotificationRepo:     notificationRepo,
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

	if err := h.ForecastingService.GenerateDemandForecast(req.ProductID, req.PeriodInDays); err != nil {
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

// GetForecastDashboard godoc
// @Summary Get forecasting dashboard data
// @Description Retrieves aggregated forecasting data including top predicted demand and low stock warnings.
// @Tags replenishment
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/dashboard [get]
func (h *ReplenishmentHandler) GetForecastDashboard(c *gin.Context) {
	dashboard, err := h.ForecastingService.GetForecastDashboard()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to get forecast dashboard", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, dashboard)
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
	if err := h.ReplenishmentService.GenerateReorderSuggestions(); err != nil {
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
	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	authUserID, ok := userIDVal.(uint)
	if !ok {
		c.Error(appErrors.NewAppError("Invalid user ID type in context", http.StatusInternalServerError, nil))
		return
	}

	var po domain.PurchaseOrder
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		var suggestion domain.ReorderSuggestion
		if err := tx.Preload("Product").Preload("Supplier").Clauses(clause.Locking{Strength: "UPDATE"}).First(&suggestion, suggestionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Reorder suggestion not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch suggestion", http.StatusInternalServerError, err)
		}

		if suggestion.Status != "PENDING" {
			return appErrors.NewAppError("Suggestion is not in PENDING state", http.StatusBadRequest, nil)
		}

		po = domain.PurchaseOrder{
			SupplierID: suggestion.SupplierID,
			Status:     "DRAFT",
			OrderDate:  time.Now(),
			CreatedBy:  authUserID,
			PurchaseOrderItems: []domain.PurchaseOrderItem{
				{
					ProductID:       suggestion.ProductID,
					OrderedQuantity: suggestion.SuggestedOrderQuantity,
					UnitPrice:       suggestion.Product.PurchasePrice,
				},
			},
		}

		if err := tx.Create(&po).Error; err != nil {
			return appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err)
		}

		if err := tx.Model(&suggestion).Update("Status", "PO_CREATED").Error; err != nil {
			return appErrors.NewAppError("Failed to update suggestion status", http.StatusInternalServerError, err)
		}
		return nil
	})

	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
			return
		}
		c.Error(appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err))
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
	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	authUserID, ok := userIDVal.(uint)
	if !ok {
		c.Error(appErrors.NewAppError("Invalid user ID type in context", http.StatusInternalServerError, nil))
		return
	}

	var po domain.PurchaseOrder
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, poID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
		}

		if po.Status != "DRAFT" {
			return appErrors.NewAppError("Purchase Order is not in DRAFT state", http.StatusBadRequest, nil)
		}

		return tx.Model(&po).Updates(map[string]interface{}{
			"Status":     "APPROVED",
			"ApprovedBy": authUserID,
			"ApprovedAt": time.Now(),
		}).Error
	})

	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
			return
		}
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
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("PurchaseOrderItems").First(&po, poID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
		}

		if po.Status != "APPROVED" && po.Status != "SENT" {
			return appErrors.NewAppError(fmt.Sprintf("Cannot receive goods for Purchase Order in %s status", po.Status), http.StatusConflict, nil)
		}

		for _, receivedItem := range req.Items {
			var poItem domain.PurchaseOrderItem
			if err := tx.First(&poItem, receivedItem.PurchaseOrderItemID).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Purchase Order Item %d not found", receivedItem.PurchaseOrderItemID), http.StatusNotFound, err)
			}

			if poItem.PurchaseOrderID != po.ID {
				return appErrors.NewAppError("Purchase Order Item does not belong to this Purchase Order", http.StatusBadRequest, nil)
			}

			if receivedItem.ReceivedQuantity > (poItem.OrderedQuantity - poItem.ReceivedQuantity) {
				return appErrors.NewAppError(fmt.Sprintf("Received quantity %d for item %d exceeds remaining ordered quantity %d", receivedItem.ReceivedQuantity, poItem.ID, (poItem.OrderedQuantity-poItem.ReceivedQuantity)), http.StatusBadRequest, nil)
			}

			poItem.ReceivedQuantity += receivedItem.ReceivedQuantity
			if err := tx.Save(&poItem).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to update received quantity for PO item %d", poItem.ID), http.StatusInternalServerError, err)
			}

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

		var refreshedItems []domain.PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", po.ID).Find(&refreshedItems).Error; err != nil {
			return appErrors.NewAppError("Failed to reload purchase order items", http.StatusInternalServerError, err)
		}

		allReceived := true
		for _, item := range refreshedItems {
			if item.OrderedQuantity != item.ReceivedQuantity {
				allReceived = false
				break
			}
		}

		if allReceived {
			return tx.Model(&po).Updates(map[string]interface{}{"Status": "RECEIVED", "ActualDeliveryDate": time.Now()}).Error
		}

		return tx.Model(&po).Update("Status", "PARTIALLY_RECEIVED").Error
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

// CreatePurchaseOrder godoc
// @Summary Create a new Purchase Order
// @Description Creates a new Purchase Order manually
// @Tags replenishment
// @Accept json
// @Produce json
// @Param request body requests.CreatePORequest true "Create PO request"
// @Success 201 {object} domain.PurchaseOrder
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/purchase-orders [post]
func (h *ReplenishmentHandler) CreatePurchaseOrder(c *gin.Context) {
	var req requests.CreatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	authUserID, ok := userIDVal.(uint)
	if !ok {
		c.Error(appErrors.NewAppError("Invalid user ID type in context", http.StatusInternalServerError, nil))
		return
	}

	var po domain.PurchaseOrder
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		po = domain.PurchaseOrder{
			SupplierID:           req.SupplierID,
			Status:               "DRAFT",
			OrderDate:            req.OrderDate,
			ExpectedDeliveryDate: req.ExpectedDeliveryDate,
			CreatedBy:            authUserID,
		}

		if err := tx.Create(&po).Error; err != nil {
			return appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err)
		}

		for _, itemReq := range req.PurchaseOrderItems {
			poItem := domain.PurchaseOrderItem{
				PurchaseOrderID: po.ID,
				ProductID:       itemReq.ProductID,
				OrderedQuantity: itemReq.OrderedQuantity,
				UnitPrice:       itemReq.UnitPrice,
			}
			if err := tx.Create(&poItem).Error; err != nil {
				return appErrors.NewAppError("Failed to create PO item", http.StatusInternalServerError, err)
			}
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
			return
		}
		c.Error(appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err))
		return
	}

	// Reload PO with items
	repository.DB.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, po.ID)
	c.JSON(http.StatusCreated, po)
}

// ListPurchaseOrders godoc
// @Summary List all purchase orders
// @Description Retrieves all purchase orders
// @Tags replenishment
// @Accept json
// @Produce json
// @Success 200 {array} domain.PurchaseOrder
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/purchase-orders [get]
func (h *ReplenishmentHandler) ListPurchaseOrders(c *gin.Context) {
	var pos []domain.PurchaseOrder
	if err := repository.DB.Preload("Supplier").Preload("PurchaseOrderItems.Product").Order("created_at desc").Find(&pos).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch purchase orders", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"purchaseOrders": pos})
}

type CreatePurchaseReturnRequest struct {
	PurchaseOrderID uint   `json:"purchase_order_id"` // Optional
	SupplierID      uint   `json:"supplier_id" binding:"required"`
	Reason          string `json:"reason" binding:"required"`
	Items           []struct {
		ProductID uint   `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
		BatchID   uint   `json:"batch_id" binding:"required"`
		Reason    string `json:"reason"`
	} `json:"items" binding:"required"`
}

// CreatePurchaseReturn godoc
// @Summary Create a return to supplier (Vendor Return)
// @Description Creates a return request to a supplier and deducts stock from the specified batch.
// @Tags replenishment
// @Accept json
// @Produce json
// @Param request body CreatePurchaseReturnRequest true "Return request"
// @Success 201 {object} domain.PurchaseReturn
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/returns [post]
func (h *ReplenishmentHandler) CreatePurchaseReturn(c *gin.Context) {
	var req CreatePurchaseReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	var returnRecord domain.PurchaseReturn

	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		returnRecord = domain.PurchaseReturn{
			PurchaseOrderID: req.PurchaseOrderID,
			SupplierID:      req.SupplierID,
			Status:          "COMPLETED", // Auto-complete for now as it deducts stock immediately
			Reason:          req.Reason,
			ReturnedBy:      userID,
			ReturnedAt:      time.Now(),
		}

		if err := tx.Create(&returnRecord).Error; err != nil {
			return appErrors.NewAppError("Failed to create return record", http.StatusInternalServerError, err)
		}

		var totalRefundAmount float64

		for _, item := range req.Items {
			// 1. Validate Batch and Stock
			var batch domain.Batch
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&batch, item.BatchID).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Batch %d not found", item.BatchID), http.StatusNotFound, err)
			}
			if batch.ProductID != item.ProductID {
				return appErrors.NewAppError("Batch does not match product", http.StatusBadRequest, nil)
			}
			if batch.Quantity < item.Quantity {
				return appErrors.NewAppError(fmt.Sprintf("Insufficient quantity in batch %s", batch.BatchNumber), http.StatusBadRequest, nil)
			}

			// 2. Deduct Stock
			batch.Quantity -= item.Quantity
			if err := tx.Save(&batch).Error; err != nil {
				return appErrors.NewAppError("Failed to update batch quantity", http.StatusInternalServerError, err)
			}

			// 3. Create Return Item
			batchID := item.BatchID
			returnItem := domain.PurchaseReturnItem{
				PurchaseReturnID: returnRecord.ID,
				ProductID:        item.ProductID,
				Quantity:         item.Quantity,
				BatchID:          &batchID, // Link to specific batch
				Reason:           item.Reason,
			}
			if err := tx.Create(&returnItem).Error; err != nil {
				return appErrors.NewAppError("Failed to create return item", http.StatusInternalServerError, err)
			}

			// 4. Create Stock Adjustment Record (STOCK_OUT)
			// Need to fetch product location for record
			var product domain.Product
			tx.First(&product, item.ProductID) // Optimize: preload or fetch

			// Calculate Refund Amount (using PurchasePrice)
			totalRefundAmount += product.PurchasePrice * float64(item.Quantity)

			stockAdj := domain.StockAdjustment{
				ProductID:   item.ProductID,
				LocationID:  product.LocationID, // Or batch location
				Type:        "STOCK_OUT",
				Quantity:    item.Quantity,
				ReasonCode:  "VENDOR_RETURN",
				Notes:       fmt.Sprintf("Vendor Return ID: %d", returnRecord.ID),
				AdjustedBy:  userID,
				AdjustedAt:  time.Now(),
				NewQuantity: batch.Quantity, // Approx batch qty
			}
			if err := tx.Create(&stockAdj).Error; err != nil {
				return appErrors.NewAppError("Failed to create stock adjustment log", http.StatusInternalServerError, err)
			}
		}

		// Update total amount
		if err := tx.Model(&returnRecord).Update("refund_amount", totalRefundAmount).Error; err != nil {
			return appErrors.NewAppError("Failed to update refund amount", http.StatusInternalServerError, err)
		}

		return nil
	})

	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Transaction failed", http.StatusInternalServerError, err))
		}
		return
	}

	// Broadcast event with FULL payload to authorized users (inventory.view)
	h.Hub.BroadcastToPermission("inventory.view", gin.H{
		"event": "RETURN_UPDATED",
		"type":  "VENDOR_RETURN",
		"data":  returnRecord,
	})

	// Create persistent notification for inventory staff
	invNotif := domain.Notification{
		Type:        "VENDOR_RETURN",
		Title:       "Vendor Return Created",
		Message:     fmt.Sprintf("Return #%d to Supplier #%d has been created.", returnRecord.ID, returnRecord.SupplierID),
		TriggeredAt: time.Now(),
	}
	h.NotificationRepo.CreateNotificationsForPermission("inventory.view", invNotif)

	// Reload with items
	repository.DB.Preload("PurchaseReturnItems.Product").First(&returnRecord, returnRecord.ID)
	c.JSON(http.StatusCreated, returnRecord)
}

// ListPurchaseReturns godoc
// @Summary List vendor returns
// @Description Retrieves all returns to suppliers
// @Tags replenishment
// @Accept json
// @Produce json
// @Success 200 {array} domain.PurchaseReturn
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/returns [get]
func (h *ReplenishmentHandler) ListPurchaseReturns(c *gin.Context) {
	var returns []domain.PurchaseReturn
	if err := repository.DB.Preload("Supplier").Preload("PurchaseReturnItems.Product").Order("created_at desc").Find(&returns).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch returns", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"returns": returns})
}
