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
		&domain.AuditLog{},
		&domain.SystemSetting{},
	))
	return db
}

func setupSalesTestDBWithFile(t *testing.T) *gorm.DB {
	t.Helper()
	path := t.TempDir() + "/test.db"
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
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
		&domain.AuditLog{},
		&domain.SystemSetting{},
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
func (stubSettingsService) UpdateSetting(key, value string, _ uint) error {
	return assert.AnError
}

type mapSettingsService struct {
	values map[string]string
}

func (m mapSettingsService) GetAllSettings() (map[string][]domain.SystemSetting, error) {
	return nil, nil
}
func (m mapSettingsService) GetSetting(key string) (string, error) {
	v, ok := m.values[key]
	if !ok {
		return "", assert.AnError
	}
	return v, nil
}
func (m mapSettingsService) GetPublicSettings() (map[string]string, error) {
	return nil, nil
}
func (m mapSettingsService) UpdateSetting(key, value string, _ uint) error {
	return nil
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

func TestLoyaltyTiers_CustomThresholds(t *testing.T) {
	db := setupSalesTestDBWithFile(t)
	userID := uint(1)
	prod := seedProductWithLocation(t, db, "LOYALTY", 1)
	prod.SellingPrice = 1000
	require.NoError(t, db.Save(&prod).Error)
	expiryDate := time.Now().Add(24 * time.Hour)
	require.NoError(t, db.Create(&domain.Batch{
		ProductID:  prod.ID,
		LocationID: 1,
		Quantity:   1,
		ExpiryDate: &expiryDate,
	}).Error)
	require.NoError(t, db.Create(&domain.SystemSetting{Key: "loyalty.earning_rate", Value: "0.1"}).Error)
	require.NoError(t, db.Create(&domain.SystemSetting{Key: "loyalty.tier_gold", Value: "100"}).Error)

	svc := NewSalesService(db, stubSettingsService{}, nil)
	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items:         []requests.CheckoutItem{{ProductID: prod.ID, Quantity: 1}},
		PaymentMethod: "cash",
		CustomerID:    ptrUint(userID),
	}, userID))

	var loyalty domain.LoyaltyAccount
	require.NoError(t, db.Where("user_id = ?", userID).First(&loyalty).Error)
	assert.Equal(t, "Gold", loyalty.Tier)
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

// TestCheckout_EmitsAuditRowForLargeDiscount verifies that a checkout whose
// combined promotion+loyalty discount exceeds 10% of the pre-discount subtotal
// writes an audit_log row. This is the compliance hook for "manager-priced"
// transactions.
func TestCheckout_EmitsAuditRowForLargeDiscount(t *testing.T) {
	db := setupSalesTestDB(t)
	prod := seedProductWithLocation(t, db, "SKU3", 1)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	seedBatch(t, db, prod.ID, 1, 100, expiry)

	// 15% off the product. SellingPrice is 10, so unit price drops to 8.50;
	// discount = 1.50 × qty(2) = 3.00, subtotal before discount = 20.00,
	// discount share = 15% > 10% threshold.
	now := time.Now()
	require.NoError(t, db.Create(&domain.Promotion{
		Name:          "Big",
		DiscountType:  "PERCENTAGE",
		DiscountValue: 15,
		StartDate:     now.Add(-time.Hour),
		EndDate:       now.Add(time.Hour),
		IsActive:      true,
		ProductID:     &prod.ID,
		Priority:      1,
	}).Error)

	svc := NewSalesService(db, stubSettingsService{}, nil)

	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: prod.ID, Quantity: 2},
		},
		PaymentMethod: "cash",
	}, 42))

	var logs []domain.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1, "exactly one audit row expected for a 15%% discount")
	assert.Equal(t, "DISCOUNT", logs[0].Action)
	assert.Equal(t, "Order", logs[0].Entity)
	assert.Equal(t, uint(42), logs[0].UserID)
	assert.NotEmpty(t, logs[0].EntityID, "order number should be recorded")
	assert.NotEmpty(t, logs[0].Changes, "changes JSON should describe the discount")
	assert.False(t, logs[0].Timestamp.IsZero())
}

// TestCheckout_NoAuditRowForSmallDiscount verifies that the audit emission is
// gated on the 10% threshold: a 5% promotion must NOT write an audit row.
func TestCheckout_NoAuditRowForSmallDiscount(t *testing.T) {
	db := setupSalesTestDB(t)
	prod := seedProductWithLocation(t, db, "SKU4", 1)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	seedBatch(t, db, prod.ID, 1, 100, expiry)

	now := time.Now()
	require.NoError(t, db.Create(&domain.Promotion{
		Name:          "Small",
		DiscountType:  "PERCENTAGE",
		DiscountValue: 5,
		StartDate:     now.Add(-time.Hour),
		EndDate:       now.Add(time.Hour),
		IsActive:      true,
		ProductID:     &prod.ID,
		Priority:      1,
	}).Error)

	svc := NewSalesService(db, stubSettingsService{}, nil)

	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: prod.ID, Quantity: 2},
		},
		PaymentMethod: "cash",
	}, 42))

	var count int64
	require.NoError(t, db.Model(&domain.AuditLog{}).Count(&count).Error)
	assert.Equal(t, int64(0), count, "no audit row expected for a 5%% discount (below threshold)")
}

// TestCheckout_PersistsTaxAmount verifies that a checkout with a configured
// tax rate writes Subtotal, TaxAmount, and TotalAmount on the Order row.
// Per P3.T7, the per-order tax amount is captured at checkout so reports
// no longer have to derive tax from total/(1+rate).
func TestCheckout_PersistsTaxAmount(t *testing.T) {
	db := setupSalesTestDB(t)
	prod := seedProductWithLocation(t, db, "SKU5", 1)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	seedBatch(t, db, prod.ID, 1, 100, expiry)

	settings := mapSettingsService{values: map[string]string{
		"tax_rate_percentage": "10",
	}}
	svc := NewSalesService(db, settings, nil)

	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: prod.ID, Quantity: 1},
		},
		PaymentMethod: "cash",
	}, 1))

	var orders []domain.Order
	require.NoError(t, db.Find(&orders).Error)
	require.Len(t, orders, 1)

	o := orders[0]
	assert.InDelta(t, 10.0, o.Subtotal, 0.01, "Subtotal must equal pre-tax total")
	assert.InDelta(t, 1.0, o.TaxAmount, 0.01, "TaxAmount must equal 10%% of subtotal")
	assert.InDelta(t, 11.0, o.TotalAmount, 0.01, "TotalAmount must equal subtotal + tax")
}

// TestCheckout_PersistsTaxAmount_NoTaxConfigured verifies the new fields
// default correctly when no tax rate is set (zero tax).
func TestCheckout_PersistsTaxAmount_NoTaxConfigured(t *testing.T) {
	db := setupSalesTestDB(t)
	prod := seedProductWithLocation(t, db, "SKU6", 1)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	seedBatch(t, db, prod.ID, 1, 100, expiry)

	svc := NewSalesService(db, mapSettingsService{}, nil)

	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: prod.ID, Quantity: 1},
		},
		PaymentMethod: "cash",
	}, 1))

	var orders []domain.Order
	require.NoError(t, db.Find(&orders).Error)
	require.Len(t, orders, 1)

	o := orders[0]
	assert.InDelta(t, 10.0, o.Subtotal, 0.01)
	assert.InDelta(t, 0.0, o.TaxAmount, 0.01)
	assert.InDelta(t, 10.0, o.TotalAmount, 0.01)
}
