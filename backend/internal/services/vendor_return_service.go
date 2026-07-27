package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"net/http"
)

// Vendor return status values. Approve does NOT change state — it only
// records intent (ApprovedBy/ApprovedAt). Complete flips PENDING → COMPLETED
// and is the only transition that mutates stock. Cancel flips PENDING → CANCELLED.
const (
	vrStatusPending   = "PENDING"
	vrStatusCompleted = "COMPLETED"
	vrStatusCancelled = "CANCELLED"
)

type VendorReturnService interface {
	CreatePurchaseReturn(req CreatePurchaseReturnRequest, userID uint) (*domain.PurchaseReturn, error)
	ListPurchaseReturns() ([]domain.PurchaseReturn, error)
	GetPurchaseReturn(ctx context.Context, id uint) (*domain.PurchaseReturn, error)
	ApprovePurchaseReturn(ctx context.Context, id uint, adminID uint) error
	CompletePurchaseReturn(ctx context.Context, id uint) error
	CancelPurchaseReturn(ctx context.Context, id uint) error
}

type vendorReturnService struct {
	db *gorm.DB
}

func NewVendorReturnService(db *gorm.DB) VendorReturnService {
	return &vendorReturnService{db: db}
}

type CreatePurchaseReturnRequest struct {
	PurchaseOrderID uint   `json:"purchase_order_id"`
	SupplierID      uint   `json:"supplier_id"`
	Reason          string `json:"reason"`
	Items           []struct {
		ProductID uint   `json:"product_id"`
		Quantity  int    `json:"quantity"`
		BatchID   uint   `json:"batch_id"`
		Reason    string `json:"reason"`
	} `json:"items"`
}

// CreatePurchaseReturn creates a PENDING vendor return record. No stock
// movement happens until the record is Completed by an admin.
func (s *vendorReturnService) CreatePurchaseReturn(req CreatePurchaseReturnRequest, userID uint) (*domain.PurchaseReturn, error) {
	var returnRecord domain.PurchaseReturn

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			var batch domain.Batch
			if err := tx.First(&batch, item.BatchID).Error; err != nil {
				return fmt.Errorf("batch %d lookup: %w", item.BatchID, err)
			}
			if batch.ProductID != item.ProductID {
				return fmt.Errorf("batch does not belong to product %d (batch %d)", item.ProductID, item.BatchID)
			}
		}

		returnRecord = domain.PurchaseReturn{
			PurchaseOrderID: req.PurchaseOrderID,
			SupplierID:      req.SupplierID,
			Status:          vrStatusPending,
			Reason:          req.Reason,
			ReturnedBy:      userID,
			ReturnedAt:      time.Now(),
		}

		if err := tx.Create(&returnRecord).Error; err != nil {
			return appErrors.NewAppError("Failed to create return record", http.StatusInternalServerError, err)
		}

		for _, item := range req.Items {
			batchID := item.BatchID
			returnItem := domain.PurchaseReturnItem{
				PurchaseReturnID: returnRecord.ID,
				ProductID:        item.ProductID,
				Quantity:         item.Quantity,
				BatchID:          &batchID,
				Reason:           item.Reason,
			}
			if err := tx.Create(&returnItem).Error; err != nil {
				return appErrors.NewAppError("Failed to create return item", http.StatusInternalServerError, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &returnRecord, nil
}

// ApprovePurchaseReturn records approval intent. It does NOT change the
// return's status — the record remains PENDING. Complete is the only
// transition that mutates state and stock.
func (s *vendorReturnService) ApprovePurchaseReturn(ctx context.Context, id uint, adminID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ret domain.PurchaseReturn
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ret, id).Error; err != nil {
			return appErrors.NewAppError(fmt.Sprintf("Vendor return %d not found", id), http.StatusNotFound, err)
		}
		if ret.Status != vrStatusPending {
			return appErrors.NewAppError(
				fmt.Sprintf("cannot approve vendor return in status %s", ret.Status),
				http.StatusConflict, nil)
		}
		now := time.Now()
		ret.ApprovedBy = &adminID
		ret.ApprovedAt = &now
		return tx.Save(&ret).Error
	})
}

// CompletePurchaseReturn transitions PENDING → COMPLETED. It requires that
// Approve has been recorded first. On success it deducts stock from the
// referenced batches, writes a STOCK_OUT audit row per item, and stamps the
// estimated refund amount.
func (s *vendorReturnService) CompletePurchaseReturn(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ret domain.PurchaseReturn
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("PurchaseReturnItems").First(&ret, id).Error; err != nil {
			return appErrors.NewAppError(fmt.Sprintf("Vendor return %d not found", id), http.StatusNotFound, err)
		}
		if ret.Status != vrStatusPending {
			return appErrors.NewAppError(
				fmt.Sprintf("cannot complete vendor return in status %s", ret.Status),
				http.StatusConflict, nil)
		}
		if ret.ApprovedAt == nil {
			return appErrors.NewAppError(
				"vendor return must be approved before completion",
				http.StatusConflict, nil)
		}

		var totalRefund float64

		for _, item := range ret.PurchaseReturnItems {
			if item.BatchID == nil {
				return appErrors.NewAppError(
					fmt.Sprintf("return item %d missing batch reference", item.ID),
					http.StatusBadRequest, nil)
			}

			var batch domain.Batch
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&batch, *item.BatchID).Error; err != nil {
				return appErrors.NewAppError(
					fmt.Sprintf("batch %d not found", *item.BatchID),
					http.StatusNotFound, err)
			}
			if batch.ProductID != item.ProductID {
				return appErrors.NewAppError(
					"batch does not match product",
					http.StatusBadRequest, nil)
			}
			if batch.Quantity < item.Quantity {
				return appErrors.NewAppError(
					fmt.Sprintf("insufficient quantity in batch %s", batch.BatchNumber),
					http.StatusBadRequest, nil)
			}

			batch.Quantity -= item.Quantity
			if err := tx.Save(&batch).Error; err != nil {
				return appErrors.NewAppError("failed to update batch quantity", http.StatusInternalServerError, err)
			}

			var product domain.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return appErrors.NewAppError("failed to load product for refund", http.StatusInternalServerError, err)
			}
			totalRefund += product.PurchasePrice * float64(item.Quantity)

			adj := domain.StockAdjustment{
				ProductID:   item.ProductID,
				LocationID:  product.LocationID,
				Type:        "STOCK_OUT",
				Quantity:    item.Quantity,
				ReasonCode:  "VENDOR_RETURN",
				Notes:       fmt.Sprintf("Vendor Return ID: %d", ret.ID),
				AdjustedBy:  derefUint(ret.ApprovedBy),
				AdjustedAt:  time.Now(),
				NewQuantity: batch.Quantity,
			}
			if err := tx.Create(&adj).Error; err != nil {
				return appErrors.NewAppError("failed to create stock adjustment log", http.StatusInternalServerError, err)
			}
		}

		now := time.Now()
		ret.Status = vrStatusCompleted
		ret.CompletedAt = &now
		ret.RefundAmount = totalRefund
		return tx.Save(&ret).Error
	})
}

// CancelPurchaseReturn transitions PENDING → CANCELLED. No stock changes
// because PENDING returns never deducted stock in the first place.
func (s *vendorReturnService) CancelPurchaseReturn(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ret domain.PurchaseReturn
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ret, id).Error; err != nil {
			return appErrors.NewAppError(fmt.Sprintf("Vendor return %d not found", id), http.StatusNotFound, err)
		}
		if ret.Status != vrStatusPending {
			return appErrors.NewAppError(
				fmt.Sprintf("cannot cancel vendor return in status %s", ret.Status),
				http.StatusConflict, nil)
		}
		now := time.Now()
		ret.Status = vrStatusCancelled
		ret.CancelledAt = &now
		return tx.Save(&ret).Error
	})
}

func (s *vendorReturnService) GetPurchaseReturn(ctx context.Context, id uint) (*domain.PurchaseReturn, error) {
	var ret domain.PurchaseReturn
	if err := s.db.WithContext(ctx).
		Preload("Supplier").
		Preload("PurchaseReturnItems.Product").
		First(&ret, id).Error; err != nil {
		return nil, appErrors.NewAppError(fmt.Sprintf("Vendor return %d not found", id), http.StatusNotFound, err)
	}
	return &ret, nil
}

func (s *vendorReturnService) ListPurchaseReturns() ([]domain.PurchaseReturn, error) {
	var returns []domain.PurchaseReturn
	if err := s.db.Preload("Supplier").Preload("PurchaseReturnItems.Product").Order("created_at desc").Find(&returns).Error; err != nil {
		return nil, appErrors.NewAppError("Failed to fetch returns", http.StatusInternalServerError, err)
	}
	return returns, nil
}

func derefUint(p *uint) uint {
	if p == nil {
		return 0
	}
	return *p
}
