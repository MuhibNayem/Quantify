package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

// StockAdjustedEventPayload defines the payload for stock adjustment events.
type StockAdjustedEventPayload struct {
	ProductID uint   `json:"productId"`
	Quantity  int    `json:"quantity"`
	Type      string `json:"type"`
	Reason    string `json:"reason"`
}

// CreateBatch godoc
// @Summary Add new stock with batch information
// @Description Adds a new batch of stock for a specific product
// @Tags stock
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Param batch body requests.StockInBatchRequest true "Batch creation request"
// @Success 201 {object} domain.Batch
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{productId}/stock/batches [post]
func CreateBatch(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid product ID", http.StatusBadRequest, err))
		return
	}

	var req requests.StockInBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var product domain.Product
	if err := repository.DB.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	batch := domain.Batch{
		ProductID:   uint(productID),
		BatchNumber: req.BatchNumber,
		Quantity:    req.Quantity,
		ExpiryDate:  req.ExpiryDate,
		LocationID:  product.LocationID, // Inherit from product's default location
	}

	if err := repository.DB.Create(&batch).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to create batch", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, batch)
}

// GetProductStock godoc
// @Summary Get current stock levels for a product
// @Description Retrieves current stock levels and batch breakdown for a specific product
// @Tags stock
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Success 200 {object} map[string]interface{} "Product stock details"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{productId}/stock [get]
func GetProductStock(c *gin.Context) {
	productID := c.Param("productId")
	var product domain.Product
	if err := repository.DB.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	var batches []domain.Batch
	db := repository.DB.Where("product_id = ?", productID)

	if locationID := c.Query("locationId"); locationID != "" {
		db = db.Where("location_id = ?", locationID)
	}

	if err := db.Find(&batches).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch batches", http.StatusInternalServerError, err))
		return
	}

	totalQuantity := 0
	for _, batch := range batches {
		totalQuantity += batch.Quantity
	}

	c.JSON(http.StatusOK, gin.H{
		"productId":     product.ID,
		"currentQuantity": totalQuantity,
		"batches":       batches,
	})
}

// CreateStockAdjustment godoc
// @Summary Perform a manual stock adjustment
// @Description Performs a manual stock adjustment (stock-in or stock-out) for a product
// @Tags stock
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Param adjustment body requests.StockAdjustmentRequest true "Stock adjustment request"
// @Success 201 {object} domain.StockAdjustment
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{productId}/stock/adjustments [post]
func CreateStockAdjustment(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid product ID", http.StatusBadRequest, err))
		return
	}

	var req requests.StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var product domain.Product
	if err := repository.DB.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	// Start a database transaction
	err = repository.DB.Transaction(func(tx *gorm.DB) error {
		// Get current quantity before adjustment within the transaction
		var currentQuantity int
		tx.Model(&domain.Batch{}).Where("product_id = ?", productID).Select("sum(quantity)").Row().Scan(&currentQuantity)

		adjustment := domain.StockAdjustment{
			ProductID:        uint(productID),
			LocationID:       product.LocationID, // Inherit from product's default location
			Type:             req.Type,
			Quantity:         req.Quantity,
			ReasonCode:       req.ReasonCode,
			Notes:            req.Notes,
			AdjustedBy:       1, // TODO: Replace with actual UserID from authentication context
			AdjustedAt:       time.Now(),
			PreviousQuantity: currentQuantity,
		}

		// Perform the actual stock adjustment in batches
		if req.Type == "STOCK_IN" {
			// For simplicity, create a new batch for manual stock-in
			batch := domain.Batch{
				ProductID:   uint(productID),
				LocationID:  product.LocationID, // Inherit from product's default location
				BatchNumber: "MANUAL-" + time.Now().Format("20060102150405"), // Unique batch number
				Quantity:    req.Quantity,
				ExpiryDate:  nil, // Manual stock-in might not have expiry
			}
			if err := tx.Create(&batch).Error; err != nil {
				return fmt.Errorf("failed to add stock for adjustment: %w", err)
			}
			adjustment.NewQuantity = currentQuantity + req.Quantity
		} else if req.Type == "STOCK_OUT" {
			// For simplicity, reduce from existing batches (FIFO/FEFO logic would be more complex)
			// This is a basic implementation and needs refinement for production
			if currentQuantity < req.Quantity {
				return fmt.Errorf("not enough stock for adjustment")
			}

			quantityToReduce := req.Quantity
			var batchesToUpdate []domain.Batch
			// Fetch batches ordered by creation date (FIFO) or expiry date (FEFO)
			if err := tx.Where("product_id = ? AND location_id = ?", productID, product.LocationID).Order("created_at asc").Find(&batchesToUpdate).Error; err != nil {
				return fmt.Errorf("failed to fetch batches for adjustment: %w", err)
			}

			for _, b := range batchesToUpdate {
				if quantityToReduce == 0 {
					break
				}
				if b.Quantity >= quantityToReduce {
					b.Quantity -= quantityToReduce
					quantityToReduce = 0
				} else {
					quantityToReduce -= b.Quantity
					b.Quantity = 0
				}
				if err := tx.Save(&b).Error; err != nil {
					return fmt.Errorf("failed to update batch quantity: %w", err)
				}
			}
			adjustment.NewQuantity = currentQuantity - req.Quantity
		}

		if err := tx.Create(&adjustment).Error; err != nil {
			return fmt.Errorf("failed to record stock adjustment: %w", err)
		}

		return nil
	})

	if err != nil {
		c.Error(appErrors.NewAppError("Stock adjustment failed", http.StatusInternalServerError, err))
		return
	}

	// Publish StockAdjustedEvent
	payload := StockAdjustedEventPayload{
		ProductID: uint(productID),
		Quantity:  req.Quantity,
		Type:      req.Type,
		Reason:    req.ReasonCode,
	}
	if err := message_broker.Publish("inventory", "stock.adjusted", payload); err != nil {
		logrus.Errorf("Failed to publish stock adjusted event: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Stock adjustment successful"})
}

// ListStockHistory godoc
// @Summary Get stock adjustment history for a product
// @Description Retrieves the stock adjustment history for a specific product
// @Tags stock
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Success 200 {array} domain.StockAdjustment
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{productId}/stock/history [get]
func ListStockHistory(c *gin.Context) {
	productID := c.Param("productId")
	var product domain.Product
	if err := repository.DB.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	var history []domain.StockAdjustment
	db := repository.DB.Where("product_id = ?", productID)

	if locationID := c.Query("locationId"); locationID != "" {
		db = db.Where("location_id = ?", locationID)
	}

	if err := db.Order("adjusted_at desc").Find(&history).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch stock history", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, history)
}
