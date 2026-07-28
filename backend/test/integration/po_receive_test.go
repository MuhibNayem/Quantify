//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/services"
)

type noopEmail struct{}

func (noopEmail) SendPurchaseOrderEmail(po domain.PurchaseOrder) error { return nil }

func TestPOReceiveFlow(t *testing.T) {
	db := SetupPostgres(t)
	Migrate(t, db)

	supplier := domain.Supplier{Name: "Acme Corp", Email: "acme@test.com"}
	require.NoError(t, db.Create(&supplier).Error)

	product := domain.Product{
		SKU:           "PO-SKU-1",
		Name:          "Gadget",
		SupplierID:    supplier.ID,
		PurchasePrice: 25.0,
		SellingPrice:  50.0,
		Status:        "Active",
	}
	require.NoError(t, db.Create(&product).Error)

	poSvc := services.NewPurchaseOrderService(db, noopEmail{})

	po, err := poSvc.CreatePurchaseOrder(domain.PurchaseOrder{
		SupplierID: supplier.ID,
		Status:     "DRAFT",
		OrderDate:  time.Now(),
		CreatedBy:  1,
		PurchaseOrderItems: []domain.PurchaseOrderItem{
			{ProductID: product.ID, OrderedQuantity: 100, UnitPrice: 25.0},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, po)

	poID := fmt.Sprintf("%d", po.ID)

	approved, err := poSvc.ApprovePurchaseOrder(poID, 2)
	require.NoError(t, err)
	assert.Equal(t, "APPROVED", approved.Status)

	receivedItems := []struct {
		PurchaseOrderItemID uint
		ReceivedQuantity    int
		BatchNumber         string
		ExpiryDate          *time.Time
	}{
		{PurchaseOrderItemID: po.PurchaseOrderItems[0].ID, ReceivedQuantity: 60, BatchNumber: "BATCH-RCV-1"},
	}

	partial, err := poSvc.ReceivePurchaseOrder(context.Background(), poID, 1, receivedItems)
	require.NoError(t, err)
	assert.Equal(t, "PARTIALLY_RECEIVED", partial.Status)

	receivedItems[0].ReceivedQuantity = 40
	receivedItems[0].BatchNumber = "BATCH-RCV-2"

	full, err := poSvc.ReceivePurchaseOrder(context.Background(), poID, 1, receivedItems)
	require.NoError(t, err)
	assert.Equal(t, "RECEIVED", full.Status)

	var batches []domain.Batch
	require.NoError(t, db.Where("product_id = ?", product.ID).Find(&batches).Error)
	assert.Len(t, batches, 2)

	totalQty := 0
	for _, b := range batches {
		totalQty += b.Quantity
	}
	assert.Equal(t, 100, totalQty)

	var adjustments []domain.StockAdjustment
	require.NoError(t, db.Where("product_id = ? AND reason_code = ?", product.ID, "PURCHASE").Find(&adjustments).Error)
	assert.Len(t, adjustments, 2)
}
