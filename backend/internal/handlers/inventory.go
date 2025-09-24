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
)

// CreateStockTransfer godoc
// @Summary Create a stock transfer
// @Description Create a new stock transfer between two locations
// @Tags inventory
// @Accept json
// @Produce json
// @Param transfer body requests.StockTransferRequest true "Stock transfer request"
// @Success 201 {object} domain.StockTransfer
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /inventory/transfers [post]
func CreateStockTransfer(c *gin.Context) {
	var req requests.StockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	// Get UserID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}

	transfer := domain.StockTransfer{
		ProductID:        req.ProductID,
		SourceLocationID: req.SourceLocationID,
		DestLocationID:   req.DestLocationID,
		Quantity:         req.Quantity,
		TransferredBy:    userID.(uint),
		TransferredAt:    time.Now(),
	}

	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		// Create the stock transfer record
		if err := tx.Create(&transfer).Error; err != nil {
			return err
		}

		// Create stock adjustment for source location (stock out)
		sourceAdjustment := domain.StockAdjustment{
			ProductID:      req.ProductID,
			LocationID:     req.SourceLocationID,
			Type:           "STOCK_OUT",
			Quantity:       req.Quantity,
			ReasonCode:     "TRANSFER_OUT",
			Notes:          "Stock transfer to location " + string(req.DestLocationID),
			AdjustedBy:     userID.(uint),
			AdjustedAt:     time.Now(),
		}
		if err := tx.Create(&sourceAdjustment).Error; err != nil {
			return err
		}

		// Create stock adjustment for destination location (stock in)
		destAdjustment := domain.StockAdjustment{
			ProductID:      req.ProductID,
			LocationID:     req.DestLocationID,
			Type:           "STOCK_IN",
			Quantity:       req.Quantity,
			ReasonCode:     "TRANSFER_IN",
			Notes:          "Stock transfer from location " + string(req.SourceLocationID),
			AdjustedBy:     userID.(uint),
			AdjustedAt:     time.Now(),
		}
		if err := tx.Create(&destAdjustment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.Error(appErrors.NewAppError("Failed to create stock transfer", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, transfer)
}
