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
	Quantity         int
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
		var srcBatches []domain.Batch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ? AND location_id = ? AND quantity > 0", req.ProductID, req.SourceLocationID).
			Order("expiry_date ASC NULLS LAST, created_at ASC").
			Find(&srcBatches).Error; err != nil {
			return err
		}

		remaining := req.Quantity
		for i := range srcBatches {
			if remaining <= 0 {
				break
			}
			take := srcBatches[i].Quantity
			if take > remaining {
				take = remaining
			}
			srcBatches[i].Quantity -= take
			if err := tx.Save(&srcBatches[i]).Error; err != nil {
				return err
			}
			remaining -= take
		}
		if remaining > 0 {
			return errors.New("insufficient stock at source")
		}

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
			}
			if err := tx.Create(&dstBatch).Error; err != nil {
				return err
			}
		} else {
			dstBatch.Quantity += req.Quantity
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
			AdjustedBy: req.UserID,
			AdjustedAt: now,
		}
		adjIn := adjOut
		adjIn.LocationID = req.DestLocationID
		adjIn.Type = "STOCK_IN"
		return tx.Create(&[]domain.StockAdjustment{adjOut, adjIn}).Error
	})
}
