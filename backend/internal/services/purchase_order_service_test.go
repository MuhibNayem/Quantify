package services

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func formatID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func setupPOTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Supplier{},
		&domain.Product{},
		&domain.Location{},
		&domain.Batch{},
		&domain.StockAdjustment{},
		&domain.PurchaseOrder{},
		&domain.PurchaseOrderItem{},
	))
	return db
}

func seedPurchaseOrderForReceive(t *testing.T, db *gorm.DB, status string, items []domain.PurchaseOrderItem) domain.PurchaseOrder {
	t.Helper()
	supplier := domain.Supplier{Name: "Acme"}
	require.NoError(t, db.Create(&supplier).Error)

	po := domain.PurchaseOrder{
		SupplierID: supplier.ID,
		Status:     status,
		OrderDate:  time.Now(),
		CreatedBy:  1,
	}
	require.NoError(t, db.Create(&po).Error)

	for i := range items {
		items[i].PurchaseOrderID = po.ID
		require.NoError(t, db.Create(&items[i]).Error)
	}
	po.PurchaseOrderItems = items
	return po
}

func TestPOReceive_CreatesStockAdjustment(t *testing.T) {
	db := setupPOTestDB(t)
	po := seedPurchaseOrderForReceive(t, db, "APPROVED", []domain.PurchaseOrderItem{
		{ProductID: 1, OrderedQuantity: 50, UnitPrice: 10.0},
	})

	svc := NewPurchaseOrderService(db, nil)
	expiry := time.Now().Add(180 * 24 * time.Hour)
	_, err := svc.ReceivePurchaseOrder(
		context.Background(),
		formatID(po.ID),
		1, // userID
		[]struct {
			PurchaseOrderItemID uint
			ReceivedQuantity    int
			BatchNumber         string
			ExpiryDate          *time.Time
		}{
			{PurchaseOrderItemID: po.PurchaseOrderItems[0].ID, ReceivedQuantity: 50, BatchNumber: "B1", ExpiryDate: &expiry},
		},
	)
	require.NoError(t, err)

	var count int64
	require.NoError(t, db.Model(&domain.StockAdjustment{}).
		Where("reason_code = ? AND product_id = ?", "PURCHASE", 1).
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "StockAdjustment audit row must be written for PO receipt")
}

func TestPOReceive_PartialReceive_CreatesAdjustment(t *testing.T) {
	db := setupPOTestDB(t)
	po := seedPurchaseOrderForReceive(t, db, "APPROVED", []domain.PurchaseOrderItem{
		{ProductID: 1, OrderedQuantity: 100, UnitPrice: 10.0},
	})

	svc := NewPurchaseOrderService(db, nil)
	_, err := svc.ReceivePurchaseOrder(
		context.Background(),
		formatID(po.ID),
		7, // userID
		[]struct {
			PurchaseOrderItemID uint
			ReceivedQuantity    int
			BatchNumber         string
			ExpiryDate          *time.Time
		}{
			{PurchaseOrderItemID: po.PurchaseOrderItems[0].ID, ReceivedQuantity: 30, BatchNumber: "B1"},
		},
	)
	require.NoError(t, err)

	var count int64
	require.NoError(t, db.Model(&domain.StockAdjustment{}).
		Where("reason_code = ? AND product_id = ? AND quantity = ?", "PURCHASE", 1, 30).
		Count(&count).Error)
	assert.Equal(t, int64(1), count, "partial receive must still write one adjustment per item")
}