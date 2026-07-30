package services

import (
	"context"
	"fmt"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseOrderService interface {
	CreatePOFromSuggestion(suggestionID string, userID uint) (*domain.PurchaseOrder, error)
	SendPurchaseOrder(poID string) (*domain.PurchaseOrder, error)
	ApprovePurchaseOrder(poID string, userID uint) (*domain.PurchaseOrder, error)
	GetPurchaseOrder(poID string) (*domain.PurchaseOrder, error)
	UpdatePurchaseOrder(poID string, updates map[string]interface{}, items []domain.PurchaseOrderItem) (*domain.PurchaseOrder, error)
	ReceivePurchaseOrder(ctx context.Context, poID string, userID uint, receivedItems []struct {
		PurchaseOrderItemID uint
		ReceivedQuantity    float64
		BatchNumber         string
		ExpiryDate          *time.Time
	}) (*domain.PurchaseOrder, error)
	CancelPurchaseOrder(poID string) (*domain.PurchaseOrder, error)
	CreatePurchaseOrder(req domain.PurchaseOrder) (*domain.PurchaseOrder, error)
	ListPurchaseOrders() ([]domain.PurchaseOrder, error)
}

type purchaseOrderService struct {
	db           *gorm.DB
	emailService EmailService
}

func NewPurchaseOrderService(db *gorm.DB, emailService EmailService) PurchaseOrderService {
	return &purchaseOrderService{
		db:           db,
		emailService: emailService,
	}
}

func (s *purchaseOrderService) CreatePOFromSuggestion(suggestionID string, userID uint) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var suggestion domain.ReorderSuggestion
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&suggestion, suggestionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Reorder suggestion not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch suggestion", http.StatusInternalServerError, err)
		}

		if suggestion.Status != "PENDING" {
			return appErrors.NewAppError("Suggestion is not in PENDING state", http.StatusBadRequest, nil)
		}

		var product domain.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, suggestion.ProductID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Product for suggestion not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch product for suggestion", http.StatusInternalServerError, err)
		}

		if suggestion.SnapshotPrice != product.PurchasePrice {
			logrus.WithFields(logrus.Fields{
				"suggestion_id":   suggestion.ID,
				"product_id":      product.ID,
				"snapshot_price":  suggestion.SnapshotPrice,
				"current_price":   product.PurchasePrice,
				"price_drift":     product.PurchasePrice - suggestion.SnapshotPrice,
			}).Warn("product purchase_price changed since reorder suggestion was generated; PO uses current price")
		}

		po = domain.PurchaseOrder{
			SupplierID: suggestion.SupplierID,
			Status:     "DRAFT",
			OrderDate:  time.Now(),
			CreatedBy:  userID,
			PurchaseOrderItems: []domain.PurchaseOrderItem{
				{
					ProductID:       product.ID,
					OrderedQuantity: suggestion.SuggestedOrderQuantity,
					UnitPrice:       product.PurchasePrice,
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
		return nil, err
	}
	return &po, nil
}

func (s *purchaseOrderService) SendPurchaseOrder(poID string) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	if err := s.db.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
		}
		return nil, appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
	}

	if po.Status != "APPROVED" {
		return nil, appErrors.NewAppError("Purchase Order is not in APPROVED state", http.StatusBadRequest, nil)
	}

	if err := s.db.Model(&po).Update("Status", "SENT").Error; err != nil {
		return nil, appErrors.NewAppError("Failed to mark purchase order as sent", http.StatusInternalServerError, err)
	}

	// Reload PO with Supplier and Items for email
	s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID)

	go func() {
		if err := s.emailService.SendPurchaseOrderEmail(po); err != nil {
			logrus.WithError(err).WithField("po_id", po.ID).Error("Failed to send PO email")
		}
	}()

	return &po, nil
}

func (s *purchaseOrderService) ApprovePurchaseOrder(poID string, userID uint) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
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
			"ApprovedBy": userID,
			"ApprovedAt": time.Now(),
		}).Error
	})

	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (s *purchaseOrderService) GetPurchaseOrder(poID string) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	if err := s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
		}
		return nil, appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
	}
	return &po, nil
}

func (s *purchaseOrderService) UpdatePurchaseOrder(poID string, updates map[string]interface{}, items []domain.PurchaseOrderItem) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, poID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
		}

		if po.Status != "DRAFT" {
			return appErrors.NewAppError(fmt.Sprintf("Cannot update Purchase Order in %s status", po.Status), http.StatusConflict, nil)
		}

		if len(updates) > 0 {
			if err := tx.Model(&po).Updates(updates).Error; err != nil {
				return appErrors.NewAppError("Failed to update purchase order", http.StatusInternalServerError, err)
			}
		}

		if len(items) > 0 {
			var existingItems []domain.PurchaseOrderItem
			if err := tx.Where("purchase_order_id = ?", po.ID).Find(&existingItems).Error; err != nil {
				return appErrors.NewAppError("Failed to load existing PO items", http.StatusInternalServerError, err)
			}

			existingByProduct := make(map[uint]domain.PurchaseOrderItem, len(existingItems))
			for _, ex := range existingItems {
				existingByProduct[ex.ProductID] = ex
			}

			wantedByProduct := make(map[uint]domain.PurchaseOrderItem, len(items))
			for _, it := range items {
				if _, dup := wantedByProduct[it.ProductID]; dup {
					return appErrors.NewAppError(fmt.Sprintf("Duplicate product %d in PO items", it.ProductID), http.StatusBadRequest, nil)
				}
				wantedByProduct[it.ProductID] = it
			}

			for prodID, ex := range existingByProduct {
				if _, kept := wantedByProduct[prodID]; kept {
					continue
				}
				if ex.ReceivedQuantity > 0 {
					return appErrors.NewAppError(fmt.Sprintf("Cannot remove product %d from PO: %v already received", prodID, ex.ReceivedQuantity), http.StatusConflict, nil)
				}
				if err := tx.Delete(&ex).Error; err != nil {
					return appErrors.NewAppError("Failed to remove old PO item", http.StatusInternalServerError, err)
				}
			}

			for prodID, it := range wantedByProduct {
				if ex, ok := existingByProduct[prodID]; ok {
					ex.OrderedQuantity = it.OrderedQuantity
					ex.UnitPrice = it.UnitPrice
					if err := tx.Save(&ex).Error; err != nil {
						return appErrors.NewAppError("Failed to update PO item", http.StatusInternalServerError, err)
					}
					continue
				}
				newItem := domain.PurchaseOrderItem{
					PurchaseOrderID: po.ID,
					ProductID:       prodID,
					OrderedQuantity: it.OrderedQuantity,
					UnitPrice:       it.UnitPrice,
				}
				if err := tx.Create(&newItem).Error; err != nil {
					return appErrors.NewAppError("Failed to create PO item", http.StatusInternalServerError, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID)
	return &po, nil
}

func (s *purchaseOrderService) ReceivePurchaseOrder(ctx context.Context, poID string, userID uint, receivedItems []struct {
	PurchaseOrderItemID uint
	ReceivedQuantity    float64
	BatchNumber         string
	ExpiryDate          *time.Time
}) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("PurchaseOrderItems").First(&po, poID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
		}

		if po.Status != "APPROVED" && po.Status != "SENT" {
			return appErrors.NewAppError(fmt.Sprintf("Cannot receive goods for Purchase Order in %s status", po.Status), http.StatusConflict, nil)
		}

		receivedValue := 0.0
		for _, receivedItem := range receivedItems {
			var poItem domain.PurchaseOrderItem
			if err := tx.First(&poItem, receivedItem.PurchaseOrderItemID).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Purchase Order Item %d not found", receivedItem.PurchaseOrderItemID), http.StatusNotFound, err)
			}

			if poItem.PurchaseOrderID != po.ID {
				return appErrors.NewAppError("Purchase Order Item does not belong to this Purchase Order", http.StatusBadRequest, nil)
			}

			if receivedItem.ReceivedQuantity > (poItem.OrderedQuantity - poItem.ReceivedQuantity) {
				return appErrors.NewAppError(fmt.Sprintf("Received quantity %v for item %d exceeds remaining ordered quantity %v", receivedItem.ReceivedQuantity, poItem.ID, (poItem.OrderedQuantity-poItem.ReceivedQuantity)), http.StatusBadRequest, nil)
			}

			poItem.ReceivedQuantity += receivedItem.ReceivedQuantity
			if err := tx.Save(&poItem).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to update received quantity for PO item %d", poItem.ID), http.StatusInternalServerError, err)
			}

			// Accumulate the value of goods received for ledger/AP posting.
			receivedValue += receivedItem.ReceivedQuantity * poItem.UnitPrice

			batch := domain.Batch{
				ProductID:   poItem.ProductID,
				BatchNumber: receivedItem.BatchNumber,
				Quantity:    receivedItem.ReceivedQuantity,
				ExpiryDate:  receivedItem.ExpiryDate,
			}
			if err := tx.Create(&batch).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to create batch for product %d", poItem.ProductID), http.StatusInternalServerError, err)
			}

			adj := domain.StockAdjustment{
				ProductID:  poItem.ProductID,
				Type:       "STOCK_IN",
				ReasonCode: "PURCHASE",
				Quantity:   receivedItem.ReceivedQuantity,
				AdjustedBy: userID,
				AdjustedAt: time.Now(),
			}
			if err := tx.Create(&adj).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to create stock adjustment for product %d", poItem.ProductID), http.StatusInternalServerError, err)
			}
		}

		var refreshedItems []domain.PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", po.ID).Find(&refreshedItems).Error; err != nil {
			return appErrors.NewAppError("Failed to reload purchase order items", http.StatusInternalServerError, err)
		}

		// Post the ledger entry for goods received on credit and record the
		// amount we now owe the supplier (accounts payable).
		if receivedValue > 0 {
			ledger := NewLedgerService(s.db)
			entry, err := ledger.PostPurchaseReceipt(tx, &po, receivedValue, userID)
			if err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to post purchase receipt to ledger: %v", err), http.StatusInternalServerError, err)
			}
			ap := domain.AccountsPayable{
				SupplierID:      po.SupplierID,
				PurchaseOrderID: po.ID,
				Amount:          decimal.NewFromFloat(receivedValue).Round(4),
				PaidAmount:      decimal.Zero,
				Status:          "UNPAID",
				JournalEntryID:  entry.ID,
			}
			if err := tx.Create(&ap).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to create accounts payable: %v", err), http.StatusInternalServerError, err)
			}
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
		return nil, err
	}

	s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID)
	return &po, nil
}

func (s *purchaseOrderService) CancelPurchaseOrder(poID string) (*domain.PurchaseOrder, error) {
	var po domain.PurchaseOrder
	if err := s.db.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
		}
		return nil, appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
	}

	if po.Status != "DRAFT" && po.Status != "APPROVED" {
		return nil, appErrors.NewAppError(fmt.Sprintf("Cannot cancel Purchase Order in %s status", po.Status), http.StatusConflict, nil)
	}

	if err := s.db.Model(&po).Update("Status", "CANCELLED").Error; err != nil {
		return nil, appErrors.NewAppError("Failed to cancel purchase order", http.StatusInternalServerError, err)
	}

	return &po, nil
}

func (s *purchaseOrderService) CreatePurchaseOrder(po domain.PurchaseOrder) (*domain.PurchaseOrder, error) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&po).Error; err != nil {
			return appErrors.NewAppError("Failed to create purchase order", http.StatusInternalServerError, err)
		}

		for i := range po.PurchaseOrderItems {
			po.PurchaseOrderItems[i].PurchaseOrderID = po.ID
			if err := tx.Create(&po.PurchaseOrderItems[i]).Error; err != nil {
				return appErrors.NewAppError("Failed to create PO item", http.StatusInternalServerError, err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, po.ID)
	return &po, nil
}

func (s *purchaseOrderService) ListPurchaseOrders() ([]domain.PurchaseOrder, error) {
	var pos []domain.PurchaseOrder
	if err := s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").Order("created_at desc").Find(&pos).Error; err != nil {
		return nil, appErrors.NewAppError("Failed to fetch purchase orders", http.StatusInternalServerError, err)
	}
	return pos, nil
}
