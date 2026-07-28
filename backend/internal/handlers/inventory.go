package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
)

// CreateStockTransfer godoc
// @Summary Create a stock transfer
// @Description Create a new stock transfer between two locations
// @Tags inventory
// @Accept json
// @Produce json
// @Param transfer body requests.StockTransferRequest true "Stock transfer request"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /inventory/transfers [post]
func CreateStockTransfer(c *gin.Context) {
	var req requests.StockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("User ID not found in context", http.StatusInternalServerError, nil))
		return
	}
	userIDValue, ok := userID.(uint)
	if !ok {
		c.Error(appErrors.NewAppError("Invalid user ID in context", http.StatusInternalServerError, nil))
		return
	}

	err := services.NewStockTransferService(repository.DB()).Transfer(c.Request.Context(), services.StockTransferRequest{
		ProductID:        req.ProductID,
		SourceLocationID: req.SourceLocationID,
		DestLocationID:   req.DestLocationID,
		Quantity:         req.Quantity,
		UserID:           userIDValue,
	})
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to create stock transfer", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "transferred"})
}
