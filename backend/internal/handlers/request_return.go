package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
)

type ReturnRequest struct {
	OrderNumber string `json:"order_number" binding:"required"`
	Items       []struct {
		OrderItemID uint    `json:"order_item_id" binding:"required"`
		Quantity    float64 `json:"quantity" binding:"required,gt=0"`
		Condition   string  `json:"condition"`
		Reason      string  `json:"reason" binding:"required"`
	} `json:"items" binding:"required"`
}

// RequestReturn godoc
// @Summary Request a return for an order
// @Description Creates a return request for specific items in an order
// @Tags returns
// @Accept json
// @Produce json
// @Param request body ReturnRequest true "Return request"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Return requested"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /returns/request [post]
func (h *ReturnHandler) RequestReturn(c *gin.Context) {
	var req ReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	authUserID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	userID := authUserID.(uint)

	var returnRecord domain.Return

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var order domain.Order
		if err := tx.Where("order_number = ? AND user_id = ?", req.OrderNumber, userID).First(&order).Error; err != nil {
			return fmt.Errorf("order not found or does not belong to user")
		}

		var count int64
		if err := tx.Model(&domain.Return{}).Where("order_id = ? AND status = ?", order.ID, "PENDING").Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check for existing returns: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("a return request is already pending for this order")
		}

		returnRecord = domain.Return{
			OrderID:      order.ID,
			UserID:       userID,
			Status:       "PENDING",
			Reason:       "Return Request",
			RefundAmount: 0,
		}

		if err := tx.Create(&returnRecord).Error; err != nil {
			return fmt.Errorf("failed to create return record: %w", err)
		}

		orderItemIDs := make([]uint, len(req.Items))
		for i, item := range req.Items {
			orderItemIDs[i] = item.OrderItemID
		}

		var orderItems []domain.OrderItem
		if err := tx.Where("id IN ? AND order_id = ?", orderItemIDs, order.ID).Find(&orderItems).Error; err != nil {
			return fmt.Errorf("failed to fetch order items: %w", err)
		}

		orderItemMap := make(map[uint]domain.OrderItem)
		for _, oi := range orderItems {
			orderItemMap[oi.ID] = oi
		}

		var totalRefundAmount float64
		var returnItems []domain.ReturnItem

		taxRate := 0.0
		if val, err := h.Settings.GetSetting("tax_rate_percentage"); err == nil {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				taxRate = v
			}
		}

		for _, item := range req.Items {
			orderItem, ok := orderItemMap[item.OrderItemID]
			if !ok {
				return fmt.Errorf("order item %d not found in order", item.OrderItemID)
			}

			if item.Quantity > (orderItem.Quantity - orderItem.ReturnedQty) {
				return fmt.Errorf("invalid return quantity for item %d", item.OrderItemID)
			}

			returnItems = append(returnItems, domain.ReturnItem{
				ReturnID:    returnRecord.ID,
				OrderItemID: orderItem.ID,
				ProductID:   orderItem.ProductID,
				Quantity:    item.Quantity,
				Condition:   item.Condition,
				Reason:      item.Reason,
			})

			totalRefundAmount += computeRefund(orderItem.UnitPrice, item.Quantity, taxRate/100.0)
		}

		if len(returnItems) > 0 {
			if err := tx.Create(&returnItems).Error; err != nil {
				return fmt.Errorf("failed to create return items: %w", err)
			}
		}

		returnRecord.RefundAmount = totalRefundAmount
		if err := tx.Save(&returnRecord).Error; err != nil {
			return fmt.Errorf("failed to update return amount: %w", err)
		}

		return nil
	})

	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}

	h.Hub.BroadcastToPermission("pos.view", gin.H{
		"event": "RETURN_REQUESTED",
		"type":  "CUSTOMER_RETURN",
		"data":  returnRecord,
	})
	h.Hub.BroadcastToPermission("returns.manage", gin.H{
		"event": "RETURN_REQUESTED",
		"type":  "CUSTOMER_RETURN",
		"data":  returnRecord,
	})

	managerNotif := domain.Notification{
		Type:        "RETURN_REQUESTED",
		Title:       "New Customer Return Request",
		Message:     fmt.Sprintf("Order #%s has a return request.", req.OrderNumber),
		TriggeredAt: time.Now(),
	}
	h.NotificationRepo.CreateNotificationsForPermission("returns.manage", managerNotif)

	salesNotif := domain.Notification{
		Type:        "RETURN_REQUESTED",
		Title:       "New Return Request",
		Message:     fmt.Sprintf("Return requested for Order #%s.", req.OrderNumber),
		TriggeredAt: time.Now(),
	}
	h.NotificationRepo.CreateNotificationsForPermission("pos.view", salesNotif)

	c.JSON(http.StatusCreated, gin.H{"message": "Return request submitted successfully", "return": returnRecord})
}
