package services

import (
	"fmt"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseOrderService interface {
	CreatePOFromSuggestion(suggestionID string, userID uint) (*domain.PurchaseOrder, error)
	SendPurchaseOrder(poID string) (*domain.PurchaseOrder, error)
	ApprovePurchaseOrder(poID string, userID uint) (*domain.PurchaseOrder, error)
	GetPurchaseOrder(poID string) (*domain.PurchaseOrder, error)
	UpdatePurchaseOrder(poID string, updates map[string]interface{}, items []domain.PurchaseOrderItem) (*domain.PurchaseOrder, error)
	ReceivePurchaseOrder(poID string, receivedItems []struct {
		PurchaseOrderItemID uint
		ReceivedQuantity    int
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
		if err := tx.Preload("Product").Preload("Supplier").Clauses(clause.Locking{Strength: "UPDATE"}).First(&suggestion, suggestionID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return appErrors.NewAppError("Reorder suggestion not found", http.StatusNotFound, err)
			}
			return appErrors.NewAppError("Failed to fetch suggestion", http.StatusInternalServerError, err)
		}

		if suggestion.Status != "PENDING" {
			return appErrors.NewAppError("Suggestion is not in PENDING state", http.StatusBadRequest, nil)
		}

		po = domain.PurchaseOrder{
			SupplierID: suggestion.SupplierID,
			Status:     "DRAFT",
			OrderDate:  time.Now(),
			CreatedBy:  userID,
			PurchaseOrderItems: []domain.PurchaseOrderItem{
				{
					ProductID:       suggestion.ProductID,
					OrderedQuantity: suggestion.SuggestedOrderQuantity,
					UnitPrice:       suggestion.Product.PurchasePrice,
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
			fmt.Printf("Failed to send PO email: %v\n", err)
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
	if err := s.db.First(&po, poID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.NewAppError("Purchase Order not found", http.StatusNotFound, err)
		}
		return nil, appErrors.NewAppError("Failed to fetch purchase order", http.StatusInternalServerError, err)
	}

	if po.Status != "DRAFT" {
		return nil, appErrors.NewAppError(fmt.Sprintf("Cannot update Purchase Order in %s status", po.Status), http.StatusConflict, nil)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&po).Updates(updates).Error; err != nil {
			return nil, appErrors.NewAppError("Failed to update purchase order", http.StatusInternalServerError, err)
		}
	}

	if len(items) > 0 {
		if err := s.db.Where("purchase_order_id = ?", po.ID).Delete(&domain.PurchaseOrderItem{}).Error; err != nil {
			return nil, appErrors.NewAppError("Failed to clear existing PO items", http.StatusInternalServerError, err)
		}
		for _, item := range items {
			item.PurchaseOrderID = po.ID
			if err := s.db.Create(&item).Error; err != nil {
				return nil, appErrors.NewAppError("Failed to create PO item", http.StatusInternalServerError, err)
			}
		}
	}

	s.db.Preload("Supplier").Preload("PurchaseOrderItems.Product").First(&po, poID)
	return &po, nil
}

func (s *purchaseOrderService) ReceivePurchaseOrder(poID string, receivedItems []struct {
	PurchaseOrderItemID uint
	ReceivedQuantity    int
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

		for _, receivedItem := range receivedItems {
			var poItem domain.PurchaseOrderItem
			if err := tx.First(&poItem, receivedItem.PurchaseOrderItemID).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Purchase Order Item %d not found", receivedItem.PurchaseOrderItemID), http.StatusNotFound, err)
			}

			if poItem.PurchaseOrderID != po.ID {
				return appErrors.NewAppError("Purchase Order Item does not belong to this Purchase Order", http.StatusBadRequest, nil)
			}

			if receivedItem.ReceivedQuantity > (poItem.OrderedQuantity - poItem.ReceivedQuantity) {
				return appErrors.NewAppError(fmt.Sprintf("Received quantity %d for item %d exceeds remaining ordered quantity %d", receivedItem.ReceivedQuantity, poItem.ID, (poItem.OrderedQuantity-poItem.ReceivedQuantity)), http.StatusBadRequest, nil)
			}

			poItem.ReceivedQuantity += receivedItem.ReceivedQuantity
			if err := tx.Save(&poItem).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to update received quantity for PO item %d", poItem.ID), http.StatusInternalServerError, err)
			}

			batch := domain.Batch{
				ProductID:   poItem.ProductID,
				BatchNumber: receivedItem.BatchNumber,
				Quantity:    receivedItem.ReceivedQuantity,
				ExpiryDate:  receivedItem.ExpiryDate,
			}
			if err := tx.Create(&batch).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Failed to create batch for product %d", poItem.ProductID), http.StatusInternalServerError, err)
			}
		}

		var refreshedItems []domain.PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", po.ID).Find(&refreshedItems).Error; err != nil {
			return appErrors.NewAppError("Failed to reload purchase order items", http.StatusInternalServerError, err)
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
