package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Batch{}, &domain.StockTransfer{}, &domain.StockAdjustment{}))
	return db
}

func seedBatch(t *testing.T, db *gorm.DB, productID, locationID uint, quantity float64, expiryDate time.Time) domain.Batch {
	t.Helper()
	batch := domain.Batch{
		ProductID:   productID,
		LocationID:  locationID,
		BatchNumber: "test-batch",
		Quantity:    quantity,
		ExpiryDate:  &expiryDate,
	}
	require.NoError(t, db.Create(&batch).Error)
	return batch
}

func TestStockTransfer_UpdatesBatches(t *testing.T) {
	db := setupTestDB(t)
	productID := uint(1)
	locSrc := uint(1)
	locDst := uint(2)
	srcBatch := seedBatch(t, db, productID, locSrc, 100, time.Now().Add(30*time.Hour*24))
	dstBatch := seedBatch(t, db, productID, locDst, 0, time.Now().Add(30*time.Hour*24))

	svc := NewStockTransferService(db)
	err := svc.Transfer(context.Background(), StockTransferRequest{
		ProductID: productID, SourceLocationID: locSrc, DestLocationID: locDst,
		Quantity: 25, Reason: "rebalance", UserID: 1,
	})
	require.NoError(t, err)

	var srcQty, dstQty float64
	db.Model(&domain.Batch{}).Where("id = ?", srcBatch.ID).Select("quantity").Scan(&srcQty)
	db.Model(&domain.Batch{}).Where("id = ?", dstBatch.ID).Select("quantity").Scan(&dstQty)
	assert.Equal(t, float64(75), srcQty, "source batch should be decremented")
	assert.Equal(t, float64(25), dstQty, "destination batch should be incremented")
}
