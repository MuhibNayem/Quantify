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

type VendorReturnService interface {
	CreatePurchaseReturn(req CreatePurchaseReturnRequest, userID uint) (*domain.PurchaseReturn, error)
	ListPurchaseReturns() ([]domain.PurchaseReturn, error)
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

func (s *vendorReturnService) CreatePurchaseReturn(req CreatePurchaseReturnRequest, userID uint) (*domain.PurchaseReturn, error) {
	var returnRecord domain.PurchaseReturn

	err := s.db.Transaction(func(tx *gorm.DB) error {
		returnRecord = domain.PurchaseReturn{
			PurchaseOrderID: req.PurchaseOrderID,
			SupplierID:      req.SupplierID,
			Status:          "COMPLETED",
			Reason:          req.Reason,
			ReturnedBy:      userID,
			ReturnedAt:      time.Now(),
		}

		if err := tx.Create(&returnRecord).Error; err != nil {
			return appErrors.NewAppError("Failed to create return record", http.StatusInternalServerError, err)
		}

		var totalRefundAmount float64

		for _, item := range req.Items {
			// 1. Validate Batch and Stock
			var batch domain.Batch
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&batch, item.BatchID).Error; err != nil {
				return appErrors.NewAppError(fmt.Sprintf("Batch %d not found", item.BatchID), http.StatusNotFound, err)
			}
			if batch.ProductID != item.ProductID {
				return appErrors.NewAppError("Batch does not match product", http.StatusBadRequest, nil)
			}
			if batch.Quantity < item.Quantity {
				return appErrors.NewAppError(fmt.Sprintf("Insufficient quantity in batch %s", batch.BatchNumber), http.StatusBadRequest, nil)
			}

			// 2. Deduct Stock
			batch.Quantity -= item.Quantity
			if err := tx.Save(&batch).Error; err != nil {
				return appErrors.NewAppError("Failed to update batch quantity", http.StatusInternalServerError, err)
			}

			// 3. Create Return Item
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

			// 4. Create Stock Adjustment Record (STOCK_OUT)
			var product domain.Product
			tx.First(&product, item.ProductID)

			totalRefundAmount += product.PurchasePrice * float64(item.Quantity)

			stockAdj := domain.StockAdjustment{
				ProductID:   item.ProductID,
				LocationID:  product.LocationID,
				Type:        "STOCK_OUT",
				Quantity:    item.Quantity,
				ReasonCode:  "VENDOR_RETURN",
				Notes:       fmt.Sprintf("Vendor Return ID: %d", returnRecord.ID),
				AdjustedBy:  userID,
				AdjustedAt:  time.Now(),
				NewQuantity: batch.Quantity,
			}
			if err := tx.Create(&stockAdj).Error; err != nil {
				return appErrors.NewAppError("Failed to create stock adjustment log", http.StatusInternalServerError, err)
			}
		}

		if err := tx.Model(&returnRecord).Update("refund_amount", totalRefundAmount).Error; err != nil {
			return appErrors.NewAppError("Failed to update refund amount", http.StatusInternalServerError, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &returnRecord, nil
}

func (s *vendorReturnService) ListPurchaseReturns() ([]domain.PurchaseReturn, error) {
	var returns []domain.PurchaseReturn
	if err := s.db.Preload("Supplier").Preload("PurchaseReturnItems.Product").Order("created_at desc").Find(&returns).Error; err != nil {
		return nil, appErrors.NewAppError("Failed to fetch returns", http.StatusInternalServerError, err)
	}
	return returns, nil
}
