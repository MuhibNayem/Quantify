package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
)

type SalesHandler struct {
	DB               *gorm.DB
	Settings         services.SettingsService
	ReportingService *services.ReportingService
	SalesService     *services.SalesService
}

func NewSalesHandler(db *gorm.DB, settings services.SettingsService, reportingService *services.ReportingService, salesService *services.SalesService) *SalesHandler {
	return &SalesHandler{
		DB:               db,
		Settings:         settings,
		ReportingService: reportingService,
		SalesService:     salesService,
	}
}

// Checkout godoc
// @Summary Process a sales checkout
// @Description Creates a new sale transaction and deducts stock atomically
// @Tags sales
// @Accept json
// @Produce json
// @Param checkout body requests.CheckoutRequest true "Checkout request"
// @Success 201 {object} map[string]interface{} "Sale completed"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /sales/checkout [post]
func (h *SalesHandler) Checkout(c *gin.Context) {
	var req requests.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	if len(req.Items) == 0 {
		c.Error(appErrors.NewAppError("Cart is empty", http.StatusBadRequest, nil))
		return
	}

	// Get authenticated user ID
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	userID := authUserID.(uint)

	// Delegate to SalesService
	if err := h.SalesService.ProcessCheckout(req, userID); err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}

	// Trigger Real-Time Report Updates (Async)
	go func() {
		if h.ReportingService != nil {
			h.ReportingService.NotifyReportUpdate("HOURLY_HEATMAP")
			h.ReportingService.NotifyReportUpdate("SALES_BY_EMPLOYEE")
			h.ReportingService.NotifyReportUpdate("COGS_GMROI")
			h.ReportingService.NotifyReportUpdate("TAX_LIABILITY")
			h.ReportingService.NotifyReportUpdate("CATEGORY_DRILLDOWN")
			h.ReportingService.NotifyReportUpdate("CUSTOMER_INSIGHTS")
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Checkout successful",
	})
}

// ListProducts godoc
// @Summary List products with stock for POS
// @Description Retrieves all products with their current aggregated stock quantity
// @Tags sales
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /sales/products [get]
func (h *SalesHandler) ListProducts(c *gin.Context) {
	type ProductWithStock struct {
		ID            uint    `json:"ID"`
		Name          string  `json:"Name"`
		SKU           string  `json:"SKU"`
		SellingPrice  float64 `json:"SellingPrice"`
		StockQuantity int     `json:"StockQuantity"`
		CategoryID    uint    `json:"CategoryID"`
		SubCategoryID uint    `json:"SubCategoryID"`
	}

	var results []ProductWithStock

	// Optimized query: Get products and sum their batch quantities
	// Using LEFT JOIN to ensure products with 0 stock are also returned (with NULL sum -> 0)
	query := h.DB.Table("products").
		Select("products.id, products.name, products.sku, products.selling_price, products.category_id, products.sub_category_id, COALESCE(SUM(batches.quantity), 0) as stock_quantity").
		Joins("LEFT JOIN batches ON batches.product_id = products.id").
		Where("products.deleted_at IS NULL"). // Respect soft delete
		Group("products.id, products.name, products.sku, products.selling_price, products.category_id, products.sub_category_id")

	if err := query.Scan(&results).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch products", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": results})
}

// ListOrders godoc
// @Summary List orders for the authenticated user
// @Description Retrieves all orders placed by the current user
// @Tags sales
// @Accept json
// @Produce json
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /sales/orders [get]
func (h *SalesHandler) ListOrders(c *gin.Context) {
	// Get authenticated user ID
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	userID := authUserID.(uint)

	var orders []domain.Order
	// Preload items and their products, AND Returns
	if err := h.DB.Preload("OrderItems.Product").Preload("Returns").Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch orders", http.StatusInternalServerError, err))
		return
	}

	// Compute transient fields
	for i := range orders {
		orders[i].AdjustedTotal = orders[i].TotalAmount
		for _, r := range orders[i].Returns {
			if r.Status == "PENDING" {
				orders[i].HasPendingReturn = true
			}
			if r.Status == "APPROVED" || r.Status == "COMPLETED" {
				orders[i].AdjustedTotal -= r.RefundAmount
			}
		}
		if orders[i].AdjustedTotal < 0 {
			orders[i].AdjustedTotal = 0
		}
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// ListAllOrders godoc
// @Summary List all sales orders (Admin/Manager)
// @Description Retrieves all sales orders in the system
// @Tags sales
// @Accept json
// @Produce json
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /sales/history [get]
// OrderResponseDTO reduces payload size and includes computed fields
type OrderResponseDTO struct {
	ID               uint            `json:"ID"`
	OrderNumber      string          `json:"OrderNumber"`
	OrderDate        time.Time       `json:"OrderDate"`
	TotalAmount      float64         `json:"TotalAmount"`
	Status           string          `json:"Status"`
	PaymentMethod    string          `json:"PaymentMethod"`
	User             UserSummaryDTO  `json:"User"`     // Staff
	Customer         *UserSummaryDTO `json:"Customer"` // Actual Customer
	OrderItems       []OrderItemDTO  `json:"OrderItems"`
	HasPendingReturn bool            `json:"HasPendingReturn"`
	AdjustedTotal    float64         `json:"AdjustedTotal"`
}

type UserSummaryDTO struct {
	ID        uint   `json:"ID"`
	Username  string `json:"Username"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
}

type OrderItemDTO struct {
	ID         uint           `json:"ID"`
	ProductID  uint           `json:"ProductID"`
	Product    ProductSummary `json:"Product"`
	Quantity   int            `json:"Quantity"`
	UnitPrice  float64        `json:"UnitPrice"`
	TotalPrice float64        `json:"TotalPrice"`
}

type ProductSummary struct {
	ID   uint   `json:"ID"`
	Name string `json:"Name"`
	SKU  string `json:"SKU"`
}

// @Router /sales/history [get]
func (h *SalesHandler) ListAllOrders(c *gin.Context) {
	var orders []domain.Order
	// Preload necessary associations.
	// Note: We preload Returns to calculate AdjustedTotal and Pending status. Preload Customer for display.
	if err := h.DB.Preload("OrderItems.Product").Preload("User").Preload("Customer").Preload("Returns").Order("created_at desc").Find(&orders).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch all orders", http.StatusInternalServerError, err))
		return
	}

	response := make([]OrderResponseDTO, 0, len(orders))

	for _, order := range orders {
		// Calculate Computed Fields
		adjustedTotal := order.TotalAmount
		hasPending := false

		for _, r := range order.Returns {
			if r.Status == "PENDING" {
				hasPending = true
			}
			if r.Status == "APPROVED" || r.Status == "COMPLETED" {
				adjustedTotal -= r.RefundAmount
			}
		}

		// Map OrderItems
		items := make([]OrderItemDTO, 0, len(order.OrderItems))
		for _, item := range order.OrderItems {
			items = append(items, OrderItemDTO{
				ID:        item.ID,
				ProductID: item.ProductID,
				Product: ProductSummary{
					ID:   item.Product.ID,
					Name: item.Product.Name,
					SKU:  item.Product.SKU,
				},
				Quantity:   item.Quantity,
				UnitPrice:  item.UnitPrice,
				TotalPrice: item.TotalPrice,
			})
		}

		// Map User (Staff)
		userSummary := UserSummaryDTO{
			ID:        order.User.ID,
			Username:  order.User.Username,
			FirstName: order.User.FirstName,
			LastName:  order.User.LastName,
		}

		// Map Customer
		var customerSummary *UserSummaryDTO
		if order.Customer != nil {
			customerSummary = &UserSummaryDTO{
				ID:        order.Customer.ID,
				Username:  order.Customer.Username,
				FirstName: order.Customer.FirstName,
				LastName:  order.Customer.LastName,
			}
		}

		response = append(response, OrderResponseDTO{
			ID:               order.ID,
			OrderNumber:      order.OrderNumber,
			OrderDate:        order.OrderDate,
			TotalAmount:      order.TotalAmount,
			Status:           order.Status,
			PaymentMethod:    order.PaymentMethod,
			User:             userSummary,
			Customer:         customerSummary,
			OrderItems:       items,
			HasPendingReturn: hasPending,
			AdjustedTotal:    adjustedTotal,
		})
	}

	c.JSON(http.StatusOK, gin.H{"orders": response})
}

// GetOrderByNumber godoc
// @Summary Get order details by order number
// @Description Retrieves a specific order by its unique order number
// @Tags sales
// @Accept json
// @Produce json
// @Param orderNumber path string true "Order Number"
// @Success 200 {object} domain.Order
// @Failure 404 {object} map[string]interface{} "Order not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /sales/orders/{orderNumber} [get]
func (h *SalesHandler) GetOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("orderNumber")

	var order domain.Order
	// Preload items, product details, user, AND Returns
	if err := h.DB.Preload("OrderItems.Product").Preload("User").Preload("Returns").Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Order not found", http.StatusNotFound, err))
		} else {
			c.Error(appErrors.NewAppError("Failed to fetch order", http.StatusInternalServerError, err))
		}
		return
	}

	// Compute transient fields
	order.AdjustedTotal = order.TotalAmount
	for _, r := range order.Returns {
		if r.Status == "PENDING" {
			order.HasPendingReturn = true
		}
		if r.Status == "APPROVED" || r.Status == "COMPLETED" {
			order.AdjustedTotal -= r.RefundAmount
		}
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}
