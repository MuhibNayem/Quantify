package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/requests"
)

type SalesHandler struct {
	DB *gorm.DB
}

func NewSalesHandler(db *gorm.DB) *SalesHandler {
	return &SalesHandler{DB: db}
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
		var totalAmount float64

		// 1. Process each item: Validate stock and deduct
		for _, item := range req.Items {
			// Lock product row to prevent race conditions check
			var product domain.Product
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, item.ProductID).Error; err != nil {
				return fmt.Errorf("product %d not found", item.ProductID)
			}

			// Calculate total quantity available across batches
			var currentStock int64
			if err := tx.Model(&domain.Batch{}).Where("product_id = ?", item.ProductID).Select("COALESCE(SUM(quantity), 0)").Row().Scan(&currentStock); err != nil {
				return fmt.Errorf("failed to check stock for product %d", item.ProductID)
			}

			if currentStock < int64(item.Quantity) {
				return fmt.Errorf("insufficient stock for product '%s' (Available: %d, Requested: %d)", product.Name, currentStock, item.Quantity)
			}

			// Deduct from batches (FIFO/FEFO)
			quantityToReduce := item.Quantity
			var batches []domain.Batch
			if err := tx.Where("product_id = ? AND quantity > 0", item.ProductID).Order("expiry_date asc, created_at asc").Find(&batches).Error; err != nil {
				return fmt.Errorf("failed to fetch batches for product %d", item.ProductID)
			}

			for _, batch := range batches {
				if quantityToReduce <= 0 {
					break
				}

				if batch.Quantity >= quantityToReduce {
					batch.Quantity -= quantityToReduce
					quantityToReduce = 0
				} else {
					quantityToReduce -= batch.Quantity
					batch.Quantity = 0
				}

				if err := tx.Save(&batch).Error; err != nil {
					return fmt.Errorf("failed to update batch %s", batch.BatchNumber)
				}
			}

			// Record Stock Adjustment (Stock Out)
			adjustment := domain.StockAdjustment{
				ProductID:        product.ID,
				LocationID:       product.LocationID,
				Type:             "STOCK_OUT",
				Quantity:         item.Quantity,
				ReasonCode:       "SALE",
				Notes:            fmt.Sprintf("Sale to customer ID: %d", req.CustomerID), // Simplified note
				AdjustedBy:       userID,
				AdjustedAt:       time.Now(),
				PreviousQuantity: int(currentStock),
				NewQuantity:      int(currentStock) - item.Quantity,
			}
			if err := tx.Create(&adjustment).Error; err != nil {
				return fmt.Errorf("failed to create stock adjustment record")
			}

			// Accumulate total (simplified price calculation, ideally should trust backend price or verify)
			totalAmount += product.SellingPrice * float64(item.Quantity)
		}

		// 2. Update Loyalty Points if CustomerID is provided
		if req.CustomerID != nil && *req.CustomerID > 0 {
			var loyalty domain.LoyaltyAccount
			result := tx.Where("user_id = ?", *req.CustomerID).First(&loyalty)
			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					// Create if missing (auto-enrollment)
					loyalty = domain.LoyaltyAccount{
						UserID: *req.CustomerID,
						Points: 0,
						Tier:   "Bronze",
					}
					// Verify user exists first? Assuming foreign key constraint will catch it if not.
					if err := tx.Create(&loyalty).Error; err != nil {
						return fmt.Errorf("failed to create loyalty account: %w", err)
					}
				} else {
					return fmt.Errorf("failed to fetch loyalty account: %w", result.Error)
				}
			}

			pointsEarned := int(totalAmount) // 1 point per $1 spent
			loyalty.Points += pointsEarned

			// Update Tier
			if loyalty.Points >= 10000 {
				loyalty.Tier = "Platinum"
			} else if loyalty.Points >= 2500 {
				loyalty.Tier = "Gold"
			} else if loyalty.Points >= 500 {
				loyalty.Tier = "Silver"
			}

			if err := tx.Save(&loyalty).Error; err != nil {
				return fmt.Errorf("failed to update loyalty points: %w", err)
			}
		}

		// 3. Record the Sale Transaction (Simplified domain model for now)
		// We can add a Sales/Order record here if the domain exists.
		// For now we assume StockAdjustment is the primary record for inventory content.
		// We will record a financial transaction.

		saletransaction := domain.Transaction{
			OrderID:              fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), userID),
			Amount:               int64(totalAmount), // Storing as int/cents often better, keeping float consistent with domain for now or converting? Domain says int64.
			Currency:             "USD",
			PaymentMethod:        req.PaymentMethod,
			Status:               "COMPLETED",
			GatewayTransactionID: fmt.Sprintf("GW-%d", time.Now().UnixNano()), // unique placeholder
		}

		// Fix Amount type mismatch: domain has int64 (likely cents), product has float64.
		// Converting to cents:
		saletransaction.Amount = int64(totalAmount * 100)

		if err := tx.Create(&saletransaction).Error; err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}

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
