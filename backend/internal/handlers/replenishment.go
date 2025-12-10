package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
	PurchaseOrderService services.PurchaseOrderService
	VendorReturnService  services.VendorReturnService
	Hub                  *websocket.Hub
	NotificationRepo     repository.NotificationRepository
	EmailService         services.EmailService
}

func NewReplenishmentHandler(
	forecastingService services.ForecastingService,
	replenishmentService services.ReplenishmentService,
	purchaseOrderService services.PurchaseOrderService,
	vendorReturnService services.VendorReturnService,
	hub *websocket.Hub,
	notificationRepo repository.NotificationRepository,
	emailService services.EmailService,
) *ReplenishmentHandler {
	return &ReplenishmentHandler{
		ForecastingService:   forecastingService,
		ReplenishmentService: replenishmentService,
		PurchaseOrderService: purchaseOrderService,
		VendorReturnService:  vendorReturnService,
		Hub:                  hub,
		NotificationRepo:     notificationRepo,
		EmailService:         emailService,
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

	forecast, err := h.ForecastingService.GenerateDemandForecast(req.ProductID, req.PeriodInDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Demand forecast generated successfully.",
		"forecast": forecast,
	})
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
func (h *ReplenishmentHandler) GetDemandForecast(c *gin.Context) {
	forecastID := c.Param("forecastId")
	forecast, err := h.ForecastingService.GetDemandForecastByID(forecastID)
	if err != nil {
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
	status := c.Query("status")
	supplierID := c.Query("supplierId")

	suggestions, err := h.ReplenishmentService.ListReorderSuggestions(status, supplierID)
	if err != nil {
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
func (h *ReplenishmentHandler) CreatePOFromSuggestion(c *gin.Context) {
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

	po, err := h.PurchaseOrderService.CreatePOFromSuggestion(suggestionID, authUserID)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err))
		}
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
func (h *ReplenishmentHandler) SendPurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	po, err := h.PurchaseOrderService.SendPurchaseOrder(poID)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to send purchase order", http.StatusInternalServerError, err))
		}
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
func (h *ReplenishmentHandler) ApprovePurchaseOrder(c *gin.Context) {
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

	po, err := h.PurchaseOrderService.ApprovePurchaseOrder(poID, authUserID)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to approve purchase order", http.StatusInternalServerError, err))
		}
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
func (h *ReplenishmentHandler) GetPurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	po, err := h.PurchaseOrderService.GetPurchaseOrder(poID)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err))
		}
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
func (h *ReplenishmentHandler) UpdatePurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var req requests.UpdatePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
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

	var items []domain.PurchaseOrderItem
	for _, itemReq := range req.PurchaseOrderItems {
		items = append(items, domain.PurchaseOrderItem{
			ProductID:       itemReq.ProductID,
			OrderedQuantity: itemReq.OrderedQuantity,
			UnitPrice:       itemReq.UnitPrice,
		})
	}

	po, err := h.PurchaseOrderService.UpdatePurchaseOrder(poID, updates, items)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to update purchase order", http.StatusInternalServerError, err))
		}
		return
	}
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
func (h *ReplenishmentHandler) ReceivePurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	var req requests.ReceivePORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	// Convert request items to anonymous struct expected by service
	var serviceItems []struct {
		PurchaseOrderItemID uint
		ReceivedQuantity    int
		BatchNumber         string
		ExpiryDate          *time.Time
	}
	for _, item := range req.Items {
		serviceItems = append(serviceItems, struct {
			PurchaseOrderItemID uint
			ReceivedQuantity    int
			BatchNumber         string
			ExpiryDate          *time.Time
		}{
			PurchaseOrderItemID: item.PurchaseOrderItemID,
			ReceivedQuantity:    item.ReceivedQuantity,
			BatchNumber:         item.BatchNumber,
			ExpiryDate:          item.ExpiryDate,
		})
	}

	po, err := h.PurchaseOrderService.ReceivePurchaseOrder(poID, serviceItems)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to receive purchase order goods", http.StatusInternalServerError, err))
		}
		return
	}
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
func (h *ReplenishmentHandler) CancelPurchaseOrder(c *gin.Context) {
	poID := c.Param("poId")
	po, err := h.PurchaseOrderService.CancelPurchaseOrder(poID)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to cancel purchase order", http.StatusInternalServerError, err))
		}
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

	po := domain.PurchaseOrder{
		SupplierID:           req.SupplierID,
		Status:               "DRAFT",
		OrderDate:            req.OrderDate,
		ExpectedDeliveryDate: req.ExpectedDeliveryDate,
		CreatedBy:            authUserID,
	}

	for _, itemReq := range req.PurchaseOrderItems {
		po.PurchaseOrderItems = append(po.PurchaseOrderItems, domain.PurchaseOrderItem{
			ProductID:       itemReq.ProductID,
			OrderedQuantity: itemReq.OrderedQuantity,
			UnitPrice:       itemReq.UnitPrice,
		})
	}

	createdPO, err := h.PurchaseOrderService.CreatePurchaseOrder(po)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err))
		}
		return
	}

	c.JSON(http.StatusCreated, createdPO)
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
	pos, err := h.PurchaseOrderService.ListPurchaseOrders()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch purchase orders", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"purchaseOrders": pos})
}

// CreatePurchaseReturn godoc
// @Summary Create a return to supplier (Vendor Return)
// @Description Creates a return request to a supplier and deducts stock from the specified batch.
// @Tags replenishment
// @Accept json
// @Produce json
// @Param request body services.CreatePurchaseReturnRequest true "Return request"
// @Success 201 {object} domain.PurchaseReturn
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /replenishment/returns [post]
func (h *ReplenishmentHandler) CreatePurchaseReturn(c *gin.Context) {
	var req services.CreatePurchaseReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	returnRecord, err := h.VendorReturnService.CreatePurchaseReturn(req, userID)
	if err != nil {
		if appErr, ok := err.(*appErrors.AppError); ok {
			c.Error(appErr)
		} else {
			c.Error(appErrors.NewAppError("Transaction failed", http.StatusInternalServerError, err))
		}
		return
	}

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
	returns, err := h.VendorReturnService.ListPurchaseReturns()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch returns", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"returns": returns})
}
