package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"strconv"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
)

type SalesHandler struct {
	DB               *gorm.DB
	Settings         services.SettingsService
	ReportingService *services.ReportingService
}

func NewSalesHandler(db *gorm.DB, settings services.SettingsService, reportingService *services.ReportingService) *SalesHandler {
	return &SalesHandler{DB: db, Settings: settings, ReportingService: reportingService}
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

	// Transaction: All or Nothing
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Bulk Fetch Products
		productIDs := make([]uint, len(req.Items))
		itemMap := make(map[uint]int) // ProductID -> Quantity
		for i, item := range req.Items {
			productIDs[i] = item.ProductID
			itemMap[item.ProductID] = item.Quantity
		}

		var products []domain.Product
		// Lock rows for update
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			return fmt.Errorf("failed to fetch products: %w", err)
		}

		if len(products) != len(req.Items) {
			return fmt.Errorf("some products not found")
		}

		productMap := make(map[uint]domain.Product)
		for _, p := range products {
			productMap[p.ID] = p
		}

		// 2. Bulk Fetch Batches
		var allBatches []domain.Batch
		if err := tx.Where("product_id IN ? AND quantity > 0", productIDs).Order("expiry_date asc, created_at asc").Find(&allBatches).Error; err != nil {
			return fmt.Errorf("failed to fetch batches: %w", err)
		}

		// Group batches by ProductID
		batchesByProduct := make(map[uint][]*domain.Batch)
		for i := range allBatches {
			// Use pointer to modify original slice elements if needed, but here we might need to be careful.
			// Better to append pointers to the slice elements.
			batchesByProduct[allBatches[i].ProductID] = append(batchesByProduct[allBatches[i].ProductID], &allBatches[i])
		}

		var totalAmount float64
		var orderItems []domain.OrderItem
		var stockAdjustments []domain.StockAdjustment

		// 3. Process Items
		for _, item := range req.Items {
			product, ok := productMap[item.ProductID]
			if !ok {
				return fmt.Errorf("product %d not found", item.ProductID)
			}

			requestedQty := item.Quantity
			batches := batchesByProduct[item.ProductID]

			// Calculate total available stock from fetched batches
			var availableStock int
			for _, b := range batches {
				availableStock += b.Quantity
			}

			if availableStock < requestedQty {
				return fmt.Errorf("insufficient stock for product '%s' (Available: %d, Requested: %d)", product.Name, availableStock, requestedQty)
			}

			// Deduct from batches
			qtyToReduce := requestedQty
			for _, batch := range batches {
				if qtyToReduce <= 0 {
					break
				}

				if batch.Quantity >= qtyToReduce {
					batch.Quantity -= qtyToReduce
					qtyToReduce = 0
				} else {
					qtyToReduce -= batch.Quantity
					batch.Quantity = 0
				}

				// We need to save the batch updates.
				// Since we are in a transaction, we can save them individually or collect them.
				// For simplicity and safety, saving individually here is okay as it's in-memory modified.
				if err := tx.Save(batch).Error; err != nil {
					return fmt.Errorf("failed to update batch %s", batch.BatchNumber)
				}
			}

			// Prepare Stock Adjustment
			stockAdjustments = append(stockAdjustments, domain.StockAdjustment{
				ProductID:        product.ID,
				LocationID:       product.LocationID,
				Type:             "STOCK_OUT",
				Quantity:         item.Quantity,
				ReasonCode:       "SALE",
				Notes:            fmt.Sprintf("Sale to customer ID: %d", req.CustomerID),
				AdjustedBy:       userID,
				AdjustedAt:       time.Now(),
				PreviousQuantity: availableStock,
				NewQuantity:      availableStock - item.Quantity,
			})

			// Accumulate total
			totalAmount += product.SellingPrice * float64(item.Quantity)

			// Prepare Order Item (for later creation)
			orderItems = append(orderItems, domain.OrderItem{
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				UnitPrice:  product.SellingPrice,
				TotalPrice: product.SellingPrice * float64(item.Quantity),
			})
		}

		// Bulk Create Stock Adjustments
		if len(stockAdjustments) > 0 {
			if err := tx.Create(&stockAdjustments).Error; err != nil {
				return fmt.Errorf("failed to create stock adjustments: %w", err)
			}
		}

		// Apply Tax
		taxRate := 0.0
		if val, err := h.Settings.GetSetting("tax_rate_percentage"); err == nil {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				taxRate = v / 100.0
			}
		}

		taxAmount := totalAmount * taxRate
		totalAmount += taxAmount // Update total to include tax

		var discountAmount float64
		var pointsRedeemed int

		// 4. Update Loyalty Points (Redemption & Earning)
		if req.CustomerID != nil && *req.CustomerID > 0 {
			var loyalty domain.LoyaltyAccount
			result := tx.Where("user_id = ?", *req.CustomerID).First(&loyalty)
			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					loyalty = domain.LoyaltyAccount{
						UserID: *req.CustomerID,
						Points: 0,
						Tier:   "Bronze",
					}
					if err := tx.Create(&loyalty).Error; err != nil {
						return fmt.Errorf("failed to create loyalty account: %w", err)
					}
				} else {
					return fmt.Errorf("failed to fetch loyalty account: %w", result.Error)
				}
			}

			// Handle Redemption
			if req.PointsToRedeem > 0 {
				if loyalty.Points < req.PointsToRedeem {
					return fmt.Errorf("insufficient loyalty points (Available: %d, Requested: %d)", loyalty.Points, req.PointsToRedeem)
				}

				// Get redemption rate
				redemptionRate := 0.01 // Default $0.01 per point
				if val, err := h.Settings.GetSetting("loyalty_points_redemption_rate"); err == nil {
					if v, err := strconv.ParseFloat(val, 64); err == nil {
						redemptionRate = v
					}
				}

				discountAmount = float64(req.PointsToRedeem) * redemptionRate
				if discountAmount > totalAmount {
					return fmt.Errorf("discount amount (%.2f) exceeds order total (%.2f)", discountAmount, totalAmount)
				}

				loyalty.Points -= req.PointsToRedeem
				pointsRedeemed = req.PointsToRedeem
			}

			// Handle Earning (on Net Amount)
			// Get earning rate from settings
			earningRate := 1.0
			if val, err := h.Settings.GetSetting("loyalty_points_earning_rate"); err == nil {
				if v, err := strconv.ParseFloat(val, 64); err == nil {
					earningRate = v
				}
			}

			netAmount := totalAmount - discountAmount
			if netAmount < 0 {
				netAmount = 0
			}

			pointsEarned := int(netAmount * earningRate)
			loyalty.Points += pointsEarned

			// Get tier thresholds from settings
			silverThreshold := 500
			goldThreshold := 2500
			platinumThreshold := 10000

			if val, err := h.Settings.GetSetting("loyalty_tier_silver"); err == nil {
				if v, err := strconv.Atoi(val); err == nil {
					silverThreshold = v
				}
			}
			if val, err := h.Settings.GetSetting("loyalty_tier_gold"); err == nil {
				if v, err := strconv.Atoi(val); err == nil {
					goldThreshold = v
				}
			}
			if val, err := h.Settings.GetSetting("loyalty_tier_platinum"); err == nil {
				if v, err := strconv.Atoi(val); err == nil {
					platinumThreshold = v
				}
			}

			if loyalty.Points >= platinumThreshold {
				loyalty.Tier = "Platinum"
			} else if loyalty.Points >= goldThreshold {
				loyalty.Tier = "Gold"
			} else if loyalty.Points >= silverThreshold {
				loyalty.Tier = "Silver"
			}

			if err := tx.Save(&loyalty).Error; err != nil {
				return fmt.Errorf("failed to update loyalty points: %w", err)
			}
		} else if req.PointsToRedeem > 0 {
			return fmt.Errorf("cannot redeem points without a customer selected")
		}

		// 5. Create Order and Order Items
		orderNumber := fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), userID)
		order := domain.Order{
			OrderNumber:    orderNumber,
			UserID:         userID,
			TotalAmount:    totalAmount - discountAmount, // Store net amount or total? Usually total paid.
			Status:         "COMPLETED",
			PaymentMethod:  req.PaymentMethod,
			OrderDate:      time.Now(),
			PointsRedeemed: pointsRedeemed,
			DiscountAmount: discountAmount,
		}

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order record: %w", err)
		}

		// Assign OrderID to items and bulk create
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return fmt.Errorf("failed to create order items: %w", err)
		}

		// 6. Create Transaction
		saletransaction := domain.Transaction{
			OrderID:              orderNumber,
			Amount:               int64((totalAmount - discountAmount) * 100),
			Currency:             "USD",
			PaymentMethod:        req.PaymentMethod,
			Status:               "COMPLETED",
			GatewayTransactionID: fmt.Sprintf("GW-%d", time.Now().UnixNano()),
		}

		if err := tx.Create(&saletransaction).Error; err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}

		return nil
	})

	if err != nil {
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
	}

	var results []ProductWithStock

	// Optimized query: Get products and sum their batch quantities
	// Using LEFT JOIN to ensure products with 0 stock are also returned (with NULL sum -> 0)
	query := h.DB.Table("products").
		Select("products.id, products.name, products.sku, products.selling_price, COALESCE(SUM(batches.quantity), 0) as stock_quantity").
		Joins("LEFT JOIN batches ON batches.product_id = products.id").
		Where("products.deleted_at IS NULL"). // Respect soft delete
		Group("products.id, products.name, products.sku, products.selling_price")

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
func (h *SalesHandler) ListAllOrders(c *gin.Context) {
	var orders []domain.Order
	// Preload items and their products
	if err := h.DB.Preload("OrderItems.Product").Preload("User").Order("created_at desc").Find(&orders).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch all orders", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
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
