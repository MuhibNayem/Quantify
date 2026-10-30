package services

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
)

type StockTransferRequest struct {
	ProductID        uint
	SourceLocationID uint
	DestLocationID   uint
	Quantity         float64
	Reason           string
	UserID           uint
}

type StockTransferService struct {
	db *gorm.DB
}

func NewStockTransferService(db *gorm.DB) *StockTransferService {
	return &StockTransferService{db: db}
}

func (s *StockTransferService) Transfer(ctx context.Context, req StockTransferRequest) error {
	if req.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if req.SourceLocationID == req.DestLocationID {
		return errors.New("source and destination must differ")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product domain.Product
		// A missing/soft-deleted product should not block moving existing
		// batches; cost falls back to zero and batch data remains intact.
		_ = tx.Unscoped().First(&product, req.ProductID).Error

		var srcBatches []domain.Batch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ? AND location_id = ? AND quantity > 0", req.ProductID, req.SourceLocationID).
			Order("expiry_date ASC NULLS LAST, created_at ASC").
			Find(&srcBatches).Error; err != nil {
			return err
		}

		remaining := req.Quantity
		transferCost := 0.0
		for i := range srcBatches {
			if remaining <= 0 {
				break
			}
			take := srcBatches[i].Quantity
			if take > remaining {
				take = remaining
			}
			unitCost := srcBatches[i].UnitCost
			if unitCost <= 0 {
				unitCost = product.PurchasePrice
			}
			transferCost += take * unitCost

			srcBatches[i].Quantity -= take
			if err := tx.Save(&srcBatches[i]).Error; err != nil {
				return err
			}
			remaining -= take
		}
		if remaining > 0 {
			return errors.New("insufficient stock at source")
		}
		transferUnitCost := transferCost / req.Quantity

		var dstBatch domain.Batch
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ? AND location_id = ?", req.ProductID, req.DestLocationID).
			Order("created_at DESC").
			First(&dstBatch).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if dstBatch.ID == 0 {
			dstBatch = domain.Batch{
				ProductID:  req.ProductID,
				LocationID: req.DestLocationID,
				Quantity:   req.Quantity,
				UnitCost:   transferUnitCost,
			}
			if err := tx.Create(&dstBatch).Error; err != nil {
				return err
			}
		} else {
			// Weighted-average the destination batch's cost so mixed receipts
			// retain a defensible valuation.
			existingUnitCost := dstBatch.UnitCost
			if existingUnitCost <= 0 {
				existingUnitCost = product.PurchasePrice
			}
			newQty := dstBatch.Quantity + req.Quantity
			if newQty > 0 {
				dstBatch.UnitCost = (dstBatch.Quantity*existingUnitCost + transferCost) / newQty
			}
			dstBatch.Quantity = newQty
			if err := tx.Save(&dstBatch).Error; err != nil {
				return err
			}
		}

		xfer := domain.StockTransfer{
			ProductID:        req.ProductID,
			SourceLocationID: req.SourceLocationID,
			DestLocationID:   req.DestLocationID,
			Quantity:         req.Quantity,
			TransferredBy:    req.UserID,
			TransferredAt:    time.Now(),
		}
		if err := tx.Create(&xfer).Error; err != nil {
			return err
		}

		now := time.Now()
		adjOut := domain.StockAdjustment{
			ProductID:  req.ProductID,
			LocationID: req.SourceLocationID,
			Type:       "STOCK_OUT",
			ReasonCode: "TRANSFER",
			Quantity:   req.Quantity,
			UnitCost:   transferUnitCost,
			AdjustedBy: req.UserID,
			AdjustedAt: now,
		}
		adjIn := adjOut
		adjIn.LocationID = req.DestLocationID
		adjIn.Type = "STOCK_IN"
		return tx.Create(&[]domain.StockAdjustment{adjOut, adjIn}).Error
	})
}
