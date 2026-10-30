package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

func setupVendorReturnTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Supplier{},
		&domain.Product{},
		&domain.Batch{},
		&domain.StockAdjustment{},
		&domain.PurchaseOrder{},
		&domain.PurchaseReturn{},
		&domain.PurchaseReturnItem{},
		&domain.Account{},
		&domain.JournalEntry{},
		&domain.JournalLine{},
	))
	require.NoError(t, repository.SeedChartOfAccounts(db))
	return db
}

func seedVendorReturnFixtures(t *testing.T, db *gorm.DB) (supplierID, productID, batchID uint) {
	t.Helper()
	supplier := domain.Supplier{Name: "Acme Corp"}
	require.NoError(t, db.Create(&supplier).Error)

	product := domain.Product{
		SKU:           "SKU-VR-1",
		Name:          "Widget",
		PurchasePrice: 10.0,
		SellingPrice:  15.0,
		LocationID:    1,
	}
	require.NoError(t, db.Create(&product).Error)

	batch := domain.Batch{
		ProductID:   product.ID,
		LocationID:  1,
		BatchNumber: "B-VR-1",
		Quantity:    100,
		ExpiryDate:  ptrTime(time.Now().Add(180 * 24 * time.Hour)),
	}
	require.NoError(t, db.Create(&batch).Error)

	return supplier.ID, product.ID, batch.ID
}

func ptrTime(t time.Time) *time.Time { return &t }

func validVendorReturnRequest(supplierID, productID, batchID uint) CreatePurchaseReturnRequest {
	return CreatePurchaseReturnRequest{
		PurchaseOrderID: 0,
		SupplierID:      supplierID,
		Reason:          "Damaged on arrival",
		Items: []struct {
			ProductID uint    `json:"product_id"`
			Quantity  float64 `json:"quantity"`
			BatchID   uint    `json:"batch_id"`
			Reason    string  `json:"reason"`
		}{
			{ProductID: productID, Quantity: 10, BatchID: batchID, Reason: "Damaged"},
		},
	}
}

func TestVendorReturn_DefaultsPending(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)
	assert.Equal(t, "PENDING", ret.Status)

	// Stock and refund amount must NOT be touched at create time.
	var batch domain.Batch
	require.NoError(t, db.First(&batch, batchID).Error)
	assert.Equal(t, float64(100), batch.Quantity, "stock must remain unchanged while PENDING")

	var adjCount int64
	db.Model(&domain.StockAdjustment{}).Count(&adjCount)
	assert.Equal(t, int64(0), adjCount, "no stock adjustment should be recorded at create")
}

func TestVendorReturn_RejectsMismatchedBatch(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, _ := seedVendorReturnFixtures(t, db)
	product2 := domain.Product{
		SKU:           "SKU-VR-2",
		Name:          "Gadget",
		PurchasePrice: 20.0,
		SellingPrice:  30.0,
		LocationID:    1,
	}
	require.NoError(t, db.Create(&product2).Error)

	batch2 := domain.Batch{
		ProductID:   product2.ID,
		LocationID:  1,
		BatchNumber: "B-VR-2",
		Quantity:    50,
		ExpiryDate:  ptrTime(time.Now().Add(30 * 24 * time.Hour)),
	}
	require.NoError(t, db.Create(&batch2).Error)

	req := validVendorReturnRequest(supplierID, productID, batch2.ID)
	_, err := svc.CreatePurchaseReturn(req, 42)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "batch does not belong")
}

func TestVendorReturn_CompleteAfterApprove(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)

	adminID := uint(7)
	require.NoError(t, svc.ApprovePurchaseReturn(context.Background(), ret.ID, adminID))

	// Approve must NOT mutate state — status stays PENDING, stock untouched.
	var mid domain.PurchaseReturn
	require.NoError(t, db.First(&mid, ret.ID).Error)
	assert.Equal(t, "PENDING", mid.Status, "Approve must not change status")
	require.NotNil(t, mid.ApprovedBy)
	assert.Equal(t, adminID, *mid.ApprovedBy)
	require.NotNil(t, mid.ApprovedAt)

	var batchMid domain.Batch
	require.NoError(t, db.First(&batchMid, batchID).Error)
	assert.Equal(t, float64(100), batchMid.Quantity, "Approve must not touch stock")

	require.NoError(t, svc.CompletePurchaseReturn(context.Background(), ret.ID))

	final, err := svc.GetPurchaseReturn(context.Background(), ret.ID)
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", final.Status)
	require.NotNil(t, final.CompletedAt)

	// Stock must be deducted and an audit row written on completion.
	var batch domain.Batch
	require.NoError(t, db.First(&batch, batchID).Error)
	assert.Equal(t, float64(90), batch.Quantity, "stock must be deducted on Complete")

	var adj domain.StockAdjustment
	require.NoError(t, db.Where("notes LIKE ?", fmt.Sprintf("%%Vendor Return ID: %d%%", ret.ID)).
		First(&adj).Error)
	assert.Equal(t, "STOCK_OUT", adj.Type)
	assert.Equal(t, "VENDOR_RETURN", adj.ReasonCode)

	assert.InDelta(t, 100.0, final.RefundAmount, 0.01, "refund = unit price * qty")
}

func TestVendorReturn_RejectsCompleteWithoutApprove(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)

	err = svc.CompletePurchaseReturn(context.Background(), ret.ID)
	assert.Error(t, err, "Complete must refuse if Approve hasn't been recorded")

	// Status must remain PENDING and stock untouched on failed transition.
	var still domain.PurchaseReturn
	require.NoError(t, db.First(&still, ret.ID).Error)
	assert.Equal(t, "PENDING", still.Status)
}

func TestVendorReturn_CancelFromPending(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)

	require.NoError(t, svc.CancelPurchaseReturn(context.Background(), ret.ID))

	final, err := svc.GetPurchaseReturn(context.Background(), ret.ID)
	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", final.Status)
	require.NotNil(t, final.CancelledAt)

	// Cancel from PENDING must not affect stock (none was deducted).
	var batch domain.Batch
	require.NoError(t, db.First(&batch, batchID).Error)
	assert.Equal(t, float64(100), batch.Quantity)
}

func TestVendorReturn_CannotCompleteTwice(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)

	adminID := uint(7)
	require.NoError(t, svc.ApprovePurchaseReturn(context.Background(), ret.ID, adminID))
	require.NoError(t, svc.CompletePurchaseReturn(context.Background(), ret.ID))

	err = svc.CompletePurchaseReturn(context.Background(), ret.ID)
	assert.Error(t, err, "second Complete must fail")
}

func TestVendorReturn_CannotCancelCompleted(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)

	adminID := uint(7)
	require.NoError(t, svc.ApprovePurchaseReturn(context.Background(), ret.ID, adminID))
	require.NoError(t, svc.CompletePurchaseReturn(context.Background(), ret.ID))

	err = svc.CancelPurchaseReturn(context.Background(), ret.ID)
	assert.Error(t, err, "Cancel on COMPLETED must fail")
}

func TestVendorReturn_ApproveOnNonPendingFails(t *testing.T) {
	db := setupVendorReturnTestDB(t)
	svc := NewVendorReturnService(db)

	supplierID, productID, batchID := seedVendorReturnFixtures(t, db)
	ret, err := svc.CreatePurchaseReturn(validVendorReturnRequest(supplierID, productID, batchID), 42)
	require.NoError(t, err)

	adminID := uint(7)
	require.NoError(t, svc.ApprovePurchaseReturn(context.Background(), ret.ID, adminID))
	require.NoError(t, svc.CompletePurchaseReturn(context.Background(), ret.ID))

	err = svc.ApprovePurchaseReturn(context.Background(), ret.ID, adminID)
	assert.Error(t, err, "Approve after Complete must fail")
}
