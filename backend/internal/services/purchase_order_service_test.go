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
		&domain.ReorderSuggestion{},
		&domain.Account{},
		&domain.JournalEntry{},
		&domain.JournalLine{},
		&domain.AccountsPayable{},
		&domain.SupplierPayment{},
		&domain.SupplierPaymentAllocation{},
	))
	// Receiving stock posts to the ledger, so these accounts must exist.
	accounts := []domain.Account{
		{Code: domain.AccountCodeCash, Name: "Cash", Type: domain.AccountAsset},
		{Code: domain.AccountCodeInventory, Name: "Inventory", Type: domain.AccountAsset},
		{Code: domain.AccountCodeAccountsPayable, Name: "AP", Type: domain.AccountLiability},
	}
	for i := range accounts {
		require.NoError(t, db.Create(&accounts[i]).Error)
	}
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
			ReceivedQuantity    float64
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
			ReceivedQuantity    float64
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

func TestUpdatePO_RejectsNonDraftStatus(t *testing.T) {
	db := setupPOTestDB(t)
	po := seedPurchaseOrderForReceive(t, db, "APPROVED", []domain.PurchaseOrderItem{
		{ProductID: 1, OrderedQuantity: 100, UnitPrice: 10.0},
	})

	svc := NewPurchaseOrderService(db, nil)
	_, err := svc.UpdatePurchaseOrder(
		formatID(po.ID),
		nil,
		[]domain.PurchaseOrderItem{
			{ProductID: 1, OrderedQuantity: 150, UnitPrice: 10.0},
		},
	)
	require.Error(t, err, "non-DRAFT PO must not be editable")

	var item domain.PurchaseOrderItem
	require.NoError(t, db.Where("purchase_order_id = ? AND product_id = ?", po.ID, 1).First(&item).Error)
	assert.Equal(t, float64(100), item.OrderedQuantity, "APPROVED PO ordered qty must remain unchanged")
}

func TestUpdatePO_PreservesReceivedQuantity(t *testing.T) {
	db := setupPOTestDB(t)
	po := seedPurchaseOrderForReceive(t, db, "DRAFT", []domain.PurchaseOrderItem{
		{ProductID: 1, OrderedQuantity: 100, ReceivedQuantity: 30, UnitPrice: 10.0},
	})

	svc := NewPurchaseOrderService(db, nil)
	_, err := svc.UpdatePurchaseOrder(
		formatID(po.ID),
		nil,
		[]domain.PurchaseOrderItem{
			{ProductID: 1, OrderedQuantity: 120, UnitPrice: 10.0},
		},
	)
	require.NoError(t, err)

	var item domain.PurchaseOrderItem
	require.NoError(t, db.Where("purchase_order_id = ? AND product_id = ?", po.ID, 1).First(&item).Error)
	assert.Equal(t, float64(30), item.ReceivedQuantity, "received qty must be preserved")
	assert.Equal(t, float64(120), item.OrderedQuantity, "ordered qty updated")
}

func TestUpdatePO_RejectsDeleteOfReceivedItem(t *testing.T) {
	db := setupPOTestDB(t)
	po := seedPurchaseOrderForReceive(t, db, "DRAFT", []domain.PurchaseOrderItem{
		{ProductID: 1, OrderedQuantity: 100, ReceivedQuantity: 30, UnitPrice: 10.0},
		{ProductID: 2, OrderedQuantity: 50, ReceivedQuantity: 0, UnitPrice: 5.0},
	})

	svc := NewPurchaseOrderService(db, nil)
	_, err := svc.UpdatePurchaseOrder(
		formatID(po.ID),
		nil,
		[]domain.PurchaseOrderItem{
			{ProductID: 2, OrderedQuantity: 60, UnitPrice: 5.0},
		},
	)
	require.Error(t, err, "must refuse to drop product 1 which has received qty")

	var item1 domain.PurchaseOrderItem
	require.NoError(t, db.Where("purchase_order_id = ? AND product_id = ?", po.ID, 1).First(&item1).Error)
	assert.Equal(t, float64(30), item1.ReceivedQuantity, "received product must remain intact")
}

func TestPOFromSuggestion_UsesLatestPurchasePrice(t *testing.T) {
	db := setupPOTestDB(t)

	supplier := domain.Supplier{Name: "Acme"}
	require.NoError(t, db.Create(&supplier).Error)

	prod := domain.Product{
		SKU:           "SKU1",
		Name:          "Widget",
		SupplierID:    supplier.ID,
		PurchasePrice: 10.0,
		SellingPrice:  15.0,
	}
	require.NoError(t, db.Create(&prod).Error)

	sugg := domain.ReorderSuggestion{
		ProductID:              prod.ID,
		SupplierID:             supplier.ID,
		CurrentStock:           0,
		SuggestedOrderQuantity: 10,
		LeadTimeDays:           7,
		Status:                 "PENDING",
		SuggestedAt:            time.Now(),
		SnapshotPrice:          10.0,
	}
	require.NoError(t, db.Create(&sugg).Error)

	require.NoError(t, db.Model(&prod).Update("purchase_price", 12.5).Error)

	svc := NewPurchaseOrderService(db, nil)
	po, err := svc.CreatePOFromSuggestion(formatID(sugg.ID), 1)
	require.NoError(t, err)
	require.NotNil(t, po)
	require.Len(t, po.PurchaseOrderItems, 1)
	assert.Equal(t, 12.5, po.PurchaseOrderItems[0].UnitPrice, "PO must use latest product price, not the suggestion's snapshot")

	var still domain.ReorderSuggestion
	require.NoError(t, db.First(&still, sugg.ID).Error)
	assert.Equal(t, 10.0, still.SnapshotPrice, "snapshot must be preserved for audit")
	assert.Equal(t, "PO_CREATED", still.Status, "suggestion must be marked PO_CREATED")
}