package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/requests"
)

func setupSalesTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Location{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Supplier{},
		&domain.Product{},
		&domain.Batch{},
		&domain.StockAdjustment{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.Transaction{},
		&domain.LoyaltyAccount{},
		&domain.Promotion{},
		&domain.Role{},
		&domain.User{},
	))
	return db
}

type stubSettingsService struct{}

func (stubSettingsService) GetAllSettings() (map[string][]domain.SystemSetting, error) {
	return nil, assert.AnError
}
func (stubSettingsService) GetSetting(key string) (string, error) {
	return "", assert.AnError
}
func (stubSettingsService) GetPublicSettings() (map[string]string, error) {
	return nil, assert.AnError
}
func (stubSettingsService) UpdateSetting(key, value string) error {
	return assert.AnError
}

func seedProductWithLocation(t *testing.T, db *gorm.DB, sku string, defaultLocationID uint) domain.Product {
	t.Helper()
	p := domain.Product{
		SKU:          sku,
		Name:         "Test Product",
		SellingPrice: 10.0,
		LocationID:   defaultLocationID,
	}
	require.NoError(t, db.Create(&p).Error)
	return p
}

func ptrUint(v uint) *uint { return &v }

func assertBatchQty(t *testing.T, db *gorm.DB, batchID uint, want int) {
	t.Helper()
	var got int
	require.NoError(t, db.Model(&domain.Batch{}).Where("id = ?", batchID).Select("quantity").Scan(&got).Error)
	assert.Equal(t, want, got, "batch %d quantity", batchID)
}

func TestCheckout_DeductsFromRegisterLocation(t *testing.T) {
	db := setupSalesTestDB(t)
	prod := seedProductWithLocation(t, db, "SKU1", 1)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	batchLoc1 := seedBatch(t, db, prod.ID, 1, 100, expiry)
	batchLoc2 := seedBatch(t, db, prod.ID, 2, 50, expiry)

	svc := NewSalesService(db, stubSettingsService{}, nil)

	err := svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: prod.ID, Quantity: 30},
		},
		PaymentMethod:      "cash",
		RegisterLocationID: ptrUint(2),
	}, 1)
	require.NoError(t, err)

	// Only location 2's batch should be deducted; location 1 untouched.
	assertBatchQty(t, db, batchLoc2.ID, 20)
	assertBatchQty(t, db, batchLoc1.ID, 100)
}

func TestCheckout_FallsBackToProductLocationWhenRegisterLocationUnset(t *testing.T) {
	db := setupSalesTestDB(t)
	prod := seedProductWithLocation(t, db, "SKU2", 1)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	batchLoc1 := seedBatch(t, db, prod.ID, 1, 100, expiry)
	batchLoc2 := seedBatch(t, db, prod.ID, 2, 50, expiry)

	svc := NewSalesService(db, stubSettingsService{}, nil)

	err := svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: prod.ID, Quantity: 30},
		},
		PaymentMethod: "cash",
		// RegisterLocationID omitted → use product's default location (1)
	}, 1)
	require.NoError(t, err)

	// Only location 1's batch should be deducted; location 2 untouched.
	assertBatchQty(t, db, batchLoc1.ID, 70)
	assertBatchQty(t, db, batchLoc2.ID, 50)
}