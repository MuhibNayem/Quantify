//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
)

type stubSettings map[string]string

func (s stubSettings) GetAllSettings() (map[string][]domain.SystemSetting, error) { return nil, nil }
func (s stubSettings) GetSetting(key string) (string, error) {
	if v, ok := s[key]; ok {
		return v, nil
	}
	return "", nil
}
func (s stubSettings) GetPublicSettings() (map[string]string, error)    { return nil, nil }
func (s stubSettings) UpdateSetting(key, value string, userID uint) error { return nil }

func TestCheckoutReturnRefundFlow(t *testing.T) {
	db := SetupPostgres(t)
	Migrate(t, db)

	settings := stubSettings{
		"tax_rate_percentage":            "10",
		"loyalty_points_earning_rate":    "1",
		"loyalty_points_redemption_rate": "0.01",
	}

	salesSvc := services.NewSalesService(db, settings, nil)

	loc := domain.Location{Name: "Main Store"}
	require.NoError(t, db.Create(&loc).Error)

	cat := domain.Category{Name: "Electronics"}
	require.NoError(t, db.Create(&cat).Error)

	product := domain.Product{
		SKU:          "SKU-001",
		Name:         "Widget",
		CategoryID:   cat.ID,
		LocationID:   loc.ID,
		SellingPrice: 100.0,
		Status:       "Active",
	}
	require.NoError(t, db.Create(&product).Error)

	batch := domain.Batch{
		ProductID:   product.ID,
		LocationID:  loc.ID,
		BatchNumber: "BATCH-001",
		Quantity:    50,
	}
	require.NoError(t, db.Create(&batch).Error)

	customer := domain.User{
		Username: "customer1",
		Password: "pass",
		Email:    "cust1@test.com",
		IsActive: true,
	}
	require.NoError(t, db.Create(&customer).Error)

	req := requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: product.ID, Quantity: 2},
		},
		CustomerID:    &customer.ID,
		PaymentMethod: "card",
	}

	err := salesSvc.ProcessCheckout(req, 1)
	require.NoError(t, err)

	var order domain.Order
	require.NoError(t, db.Preload("OrderItems").Last(&order).Error)
	assert.Equal(t, "COMPLETED", order.Status)
	assert.Equal(t, 200.0, order.Subtotal)
	assert.Equal(t, 20.0, order.TaxAmount)
	assert.Equal(t, 220.0, order.TotalAmount)

	var batchAfter domain.Batch
	require.NoError(t, db.First(&batchAfter, batch.ID).Error)
	assert.Equal(t, 48, batchAfter.Quantity)

	returnRecord := domain.Return{
		OrderID:      order.ID,
		UserID:       customer.ID,
		Status:       "PENDING",
		Reason:       "Defective",
		RefundAmount: 110.0,
	}
	require.NoError(t, db.Create(&returnRecord).Error)

	returnItem := domain.ReturnItem{
		ReturnID:    returnRecord.ID,
		OrderItemID: order.OrderItems[0].ID,
		ProductID:   product.ID,
		Quantity:    1,
		Condition:   "DAMAGED",
		Reason:      "Defective screen",
	}
	require.NoError(t, db.Create(&returnItem).Error)

	now := time.Now()
	adminID := uint(1)
	require.NoError(t, db.Model(&returnRecord).Updates(map[string]interface{}{
		"Status":     "APPROVED",
		"ApprovedBy": adminID,
		"ApprovedAt": now,
	}).Error)

	require.NoError(t, db.Model(&returnRecord).Update("Status", "COMPLETED").Error)

	var returnAfter domain.Return
	require.NoError(t, db.First(&returnAfter, returnRecord.ID).Error)
	assert.Equal(t, "COMPLETED", returnAfter.Status)
	assert.Equal(t, 110.0, returnAfter.RefundAmount)
}
