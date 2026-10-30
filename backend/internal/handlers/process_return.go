package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/services"
	"inventory/backend/internal/units"
)

// ProcessReturn godoc
// @Summary Process a return request (Approve/Reject)
// @Description Approves or rejects a return request. If approved, updates stock and refunds.
// @Tags returns
// @Accept json
// @Produce json
// @Param id path int true "Return ID"
// @Param action body map[string]string true "Action (approve/reject)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Return processed"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /returns/{id}/process [post]
func (h *ReturnHandler) ProcessReturn(c *gin.Context) {
	returnID := c.Param("id")
	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	authUserID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	approverID := authUserID.(uint)

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var returnRecord domain.Return
		if err := tx.Preload("ReturnItems").First(&returnRecord, returnID).Error; err != nil {
			return fmt.Errorf("return request not found")
		}

		if returnRecord.Status != "PENDING" {
			return fmt.Errorf("return request is already processed")
		}

		if req.Action == "reject" {
			returnRecord.Status = "REJECTED"
			now := time.Now()
			returnRecord.ApprovedAt = &now
			returnRecord.ApprovedBy = &approverID
			return tx.Save(&returnRecord).Error
		} else if req.Action != "approve" {
			return fmt.Errorf("invalid action")
		}

		returnRecord.Status = "APPROVED"
		now := time.Now()
		returnRecord.ApprovedAt = &now
		returnRecord.ApprovedBy = &approverID

		productIDs := make([]uint, 0)
		for _, item := range returnRecord.ReturnItems {
			if item.Condition == "GOOD" {
				productIDs = append(productIDs, item.ProductID)
			}
		}

		productMap := make(map[uint]domain.Product)
		if len(productIDs) > 0 {
			var products []domain.Product
			if err := tx.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
				return fmt.Errorf("failed to fetch products: %w", err)
			}
			for _, p := range products {
				productMap[p.ID] = p
			}
		}

		var stockAdjustments []domain.StockAdjustment
		restockCost := 0.0

		for _, item := range returnRecord.ReturnItems {
			// Load the original line so the return restocks the same physical
			// quantity that was deducted (selling units → base units).
			var orderItem domain.OrderItem
			if err := tx.First(&orderItem, item.OrderItemID).Error; err != nil {
				return fmt.Errorf("order item %d not found", item.OrderItemID)
			}

			restockQty := item.Quantity
			if orderItem.Quantity > 0 && orderItem.BaseQuantity > 0 {
				restockQty = units.RoundQty(item.Quantity * orderItem.BaseQuantity / orderItem.Quantity)
			}

			if item.Condition == "GOOD" {
				product, ok := productMap[item.ProductID]
				if !ok {
					return fmt.Errorf("product %d not found", item.ProductID)
				}

				locationID := orderItem.LocationID
				if locationID == 0 {
					locationID = product.LocationID
				}
				unitCost := orderItem.UnitCost
				if unitCost <= 0 {
					unitCost = product.PurchasePrice
				}
				restockCost += restockQty * unitCost

				var batch domain.Batch
				if err := tx.Where("product_id = ? AND location_id = ?", item.ProductID, locationID).Order("created_at desc").First(&batch).Error; err == nil {
					previousQty := batch.Quantity
					batch.Quantity = units.RoundQty(batch.Quantity + restockQty)
					batch.UnitCost = unitCost
					if err := tx.Save(&batch).Error; err != nil {
						return fmt.Errorf("failed to update batch")
					}
					stockAdjustments = append(stockAdjustments, domain.StockAdjustment{
						ProductID:        item.ProductID,
						LocationID:       locationID,
						Type:             "STOCK_IN",
						Quantity:         restockQty,
						UnitCost:         unitCost,
						ReasonCode:       "RETURN",
						Notes:            fmt.Sprintf("Return ID: %d", returnRecord.ID),
						AdjustedBy:       approverID,
						AdjustedAt:       time.Now(),
						PreviousQuantity: previousQty,
						NewQuantity:      batch.Quantity,
					})
				} else {
					batch = domain.Batch{
						ProductID:   item.ProductID,
						LocationID:  locationID,
						BatchNumber: fmt.Sprintf("RET-%d", returnRecord.ID),
						Quantity:    restockQty,
						UnitCost:    unitCost,
					}
					if err := tx.Create(&batch).Error; err != nil {
						return fmt.Errorf("failed to create batch")
					}
					stockAdjustments = append(stockAdjustments, domain.StockAdjustment{
						ProductID:        item.ProductID,
						LocationID:       locationID,
						Type:             "STOCK_IN",
						Quantity:         restockQty,
						UnitCost:         unitCost,
						ReasonCode:       "RETURN",
						Notes:            fmt.Sprintf("Return ID: %d", returnRecord.ID),
						AdjustedBy:       approverID,
						AdjustedAt:       time.Now(),
						PreviousQuantity: 0,
						NewQuantity:      batch.Quantity,
					})
				}
			}

			if err := tx.Model(&domain.OrderItem{}).Where("id = ?", item.OrderItemID).
				Update("returned_qty", gorm.Expr("returned_qty + ?", item.Quantity)).
				Update("is_returned", true).Error; err != nil {
				return fmt.Errorf("failed to update order item")
			}
		}

		if len(stockAdjustments) > 0 {
			if err := tx.Create(&stockAdjustments).Error; err != nil {
				return fmt.Errorf("failed to create stock adjustments")
			}
		}

		var order domain.Order
		if err := tx.First(&order, returnRecord.OrderID).Error; err != nil {
			return fmt.Errorf("order not found")
		}

		// Split the refund into revenue and tax using the tax rate actually
		// charged on the original order, then post the return to the ledger.
		effectiveTaxRate := 0.0
		if order.Subtotal > 0 {
			effectiveTaxRate = order.TaxAmount / order.Subtotal
		}
		netRefund := returnRecord.RefundAmount
		taxRefund := 0.0
		if effectiveTaxRate > 0 {
			netRefund = returnRecord.RefundAmount / (1 + effectiveTaxRate)
			taxRefund = returnRecord.RefundAmount - netRefund
		}
		ledger := services.NewLedgerService(h.DB)
		entry, err := ledger.PostReturn(tx, &returnRecord, &order, netRefund, taxRefund, restockCost, approverID)
		if err != nil {
			return fmt.Errorf("failed to post return to ledger: %w", err)
		}
		if entry != nil {
			returnRecord.JournalEntryID = &entry.ID
		}

		// A cash refund reduces the original cashier's open drawer.
		if order.PaymentMethod == "cash" {
			var drawer domain.CashDrawerSession
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("user_id = ? AND status = ?", order.UserID, "OPEN").
				Order("start_time DESC").First(&drawer).Error; err == nil {
				drawer.TotalRefunds += returnRecord.RefundAmount
				drawer.SystemCash = drawer.StartingCash + drawer.TotalSales - drawer.TotalRefunds - drawer.TotalDrops
				if err := tx.Save(&drawer).Error; err != nil {
					return fmt.Errorf("failed to update cash drawer for refund: %w", err)
				}
			}
		}

		currency := "USD"
		if v, err := h.Settings.GetSetting("currency_code"); err == nil {
			if code := strings.ToUpper(strings.TrimSpace(v)); len(code) == 3 {
				currency = code
			}
		}

		transaction := domain.Transaction{
			OrderID:              order.OrderNumber,
			Amount:               int64(returnRecord.RefundAmount * 100),
			Currency:             currency,
			PaymentMethod:        order.PaymentMethod,
			Status:               "REFUNDED",
			GatewayTransactionID: fmt.Sprintf("REF-%d", time.Now().UnixNano()),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("failed to create refund transaction")
		}

		var orderForPoints domain.Order
		orderPointsRedeemed := 0
		if err := tx.First(&orderForPoints, returnRecord.OrderID).Error; err == nil {
			orderPointsRedeemed = orderForPoints.PointsRedeemed
		}
		totalPointsToDeduct := 0
		for _, item := range returnRecord.ReturnItems {
			var orderItem domain.OrderItem
			if err := tx.First(&orderItem, item.OrderItemID).Error; err == nil {
				totalPointsToDeduct += computeLoyaltyReversal(orderPointsRedeemed, orderItem.Quantity, item.Quantity)
			}
		}
		var loyalty domain.LoyaltyAccount
		if err := tx.Where("user_id = ?", returnRecord.UserID).First(&loyalty).Error; err == nil {
			if loyalty.Points < totalPointsToDeduct {
				return fmt.Errorf("insufficient loyalty points balance to reverse")
			}
			loyalty.Points -= totalPointsToDeduct
			if err := tx.Save(&loyalty).Error; err != nil {
				return fmt.Errorf("failed to update loyalty points")
			}
		}

		return tx.Save(&returnRecord).Error
	})

	if err != nil {
		c.Error(appErrors.NewAppError(err.Error(), http.StatusBadRequest, err))
		return
	}

	var updatedReturn domain.Return
	if err := h.DB.Preload("ReturnItems.OrderItem.Product").Preload("Order").First(&updatedReturn, returnID).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch updated return", http.StatusInternalServerError, err))
		return
	}

	h.Hub.BroadcastToPermission("returns.manage", gin.H{
		"event": "RETURN_UPDATED",
		"type":  "CUSTOMER_RETURN",
		"data":  updatedReturn,
	})

	managerNotif := domain.Notification{
		Type:        "RETURN_UPDATED",
		Title:       "Return Processed",
		Message:     fmt.Sprintf("Return #%d has been %s.", updatedReturn.ID, updatedReturn.Status),
		TriggeredAt: time.Now(),
	}
	h.NotificationRepo.CreateNotificationsForPermission("returns.manage", managerNotif)

	salesNotif := domain.Notification{
		Type:        "RETURN_UPDATED",
		Title:       "Return Update",
		Message:     fmt.Sprintf("Return #%d for Order #%s is now %s.", updatedReturn.ID, updatedReturn.Order.OrderNumber, updatedReturn.Status),
		TriggeredAt: time.Now(),
	}
	h.NotificationRepo.CreateNotificationsForPermission("pos.view", salesNotif)

	go func() {
		if h.ReportingService != nil {
			h.ReportingService.NotifyReportUpdate("RETURNS_ANALYSIS")
			h.ReportingService.NotifyReportUpdate("SHRINKAGE")
			h.ReportingService.NotifyReportUpdate("COGS_GMROI")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Return processed successfully", "return": updatedReturn})
}
