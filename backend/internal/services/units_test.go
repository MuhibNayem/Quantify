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
	"inventory/backend/internal/units"
)

func setupUnitsTestDB(t *testing.T) *gorm.DB {
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
		&domain.UnitDimension{},
		&domain.Unit{},
		&domain.ProductUnit{},
	))
	return db
}

// seedWireProduct creates a length-dimension product stocked by the foot and
// sold by the inch (inch ProductUnit factor = 1/12). Returns the product and
// the inch unit ID.
func seedWireProduct(t *testing.T, db *gorm.DB) (domain.Product, uint) {
	t.Helper()

	dim := domain.UnitDimension{Code: "length", Name: "Length"}
	require.NoError(t, db.Create(&dim).Error)

	foot := domain.Unit{Code: "ft", Name: "Foot", Symbol: "ft", DimensionID: dim.ID, Factor: 1, IsBase: true}
	require.NoError(t, db.Create(&foot).Error)

	inch := domain.Unit{Code: "in", Name: "Inch", Symbol: "in", DimensionID: dim.ID, Factor: 1.0 / 12.0}
	require.NoError(t, db.Create(&inch).Error)

	product := domain.Product{
		SKU:             "WIRE",
		Name:            "Copper Wire",
		SellingPrice:    0.50, // per foot (base unit)
		LocationID:      1,
		BaseUnitID:      &foot.ID,
		AllowFractional: true,
	}
	require.NoError(t, db.Create(&product).Error)

	pu := domain.ProductUnit{
		ProductID: product.ID,
		UnitID:    inch.ID,
		Factor:    1.0 / 12.0,
		IsActive:  true,
	}
	require.NoError(t, db.Create(&pu).Error)

	return product, inch.ID
}

// TestCheckout_UnitAwareStockDeduction verifies that selling 14 inches of a
// foot-stocked product deducts 14 * (1/12) = 1.1667 ft from the base-unit batch
// and records the selling-unit quantity plus base-unit equivalent on the order.
func TestCheckout_UnitAwareStockDeduction(t *testing.T) {
	db := setupUnitsTestDB(t)
	product, inchID := seedWireProduct(t, db)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	batch := seedBatch(t, db, product.ID, 1, 100, expiry)

	svc := NewSalesService(db, stubSettingsService{}, nil)
	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: product.ID, Quantity: 14, UnitID: &inchID},
		},
		PaymentMethod: "cash",
	}, 1))

	assertBatchQty(t, db, batch.ID, units.RoundQty(100-1.1667))

	var items []domain.OrderItem
	require.NoError(t, db.Preload("Unit").Find(&items).Error)
	require.Len(t, items, 1)

	item := items[0]
	assert.InDelta(t, 14, item.Quantity, 0.0001, "selling-unit quantity preserved")
	require.NotNil(t, item.UnitID, "selling unit recorded")
	assert.Equal(t, inchID, *item.UnitID)
	assert.InDelta(t, 1.1667, item.BaseQuantity, 0.0001, "base-unit equivalent recorded")
	// No explicit inch price → derived: 0.50/ft × (1/12) per inch.
	assert.InDelta(t, units.RoundQty(0.50*(1.0/12.0)), item.UnitPrice, 0.0001)
	assert.InDelta(t, units.RoundQty(units.RoundQty(0.50*(1.0/12.0))*14), item.TotalPrice, 0.0001)
}

// TestCheckout_BaseUnitWhenUnitOmitted verifies that omitting UnitID sells in
// the product's base unit with a factor of 1 (no conversion).
func TestCheckout_BaseUnitWhenUnitOmitted(t *testing.T) {
	db := setupUnitsTestDB(t)
	product, _ := seedWireProduct(t, db)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	batch := seedBatch(t, db, product.ID, 1, 100, expiry)

	svc := NewSalesService(db, stubSettingsService{}, nil)
	require.NoError(t, svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: product.ID, Quantity: 1.2},
		},
		PaymentMethod: "cash",
	}, 1))

	assertBatchQty(t, db, batch.ID, units.RoundQty(100-1.2))

	var items []domain.OrderItem
	require.NoError(t, db.Find(&items).Error)
	require.Len(t, items, 1)
	assert.Nil(t, items[0].UnitID)
	assert.InDelta(t, 1.2, items[0].BaseQuantity, 0.0001)
	assert.InDelta(t, 0.50, items[0].UnitPrice, 0.0001)
}

// TestCheckout_FractionalGuard verifies that a non-fractional product rejects a
// fractional selling quantity.
func TestCheckout_FractionalGuard(t *testing.T) {
	db := setupUnitsTestDB(t)

	product := domain.Product{
		SKU:             "EACH",
		Name:            "Widget",
		SellingPrice:    5.0,
		LocationID:      1,
		AllowFractional: false,
	}
	require.NoError(t, db.Create(&product).Error)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	seedBatch(t, db, product.ID, 1, 100, expiry)

	svc := NewSalesService(db, stubSettingsService{}, nil)
	err := svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: product.ID, Quantity: 1.5},
		},
		PaymentMethod: "cash",
	}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fractional")
}

// TestCheckout_IncompatibleUnitRejected verifies that a selling unit from a
// different dimension than the product's base unit is rejected.
func TestCheckout_IncompatibleUnitRejected(t *testing.T) {
	db := setupUnitsTestDB(t)
	product, _ := seedWireProduct(t, db)

	// A mass unit (kg) sharing no dimension with the length-based wire.
	mass := domain.UnitDimension{Code: "mass", Name: "Mass"}
	require.NoError(t, db.Create(&mass).Error)
	kg := domain.Unit{Code: "kg", Name: "Kilogram", Symbol: "kg", DimensionID: mass.ID, Factor: 1, IsBase: true}
	require.NoError(t, db.Create(&kg).Error)
	require.NoError(t, db.Create(&domain.ProductUnit{
		ProductID: product.ID,
		UnitID:    kg.ID,
		Factor:    1,
		IsActive:  true,
	}).Error)

	expiry := time.Now().Add(180 * 24 * time.Hour)
	seedBatch(t, db, product.ID, 1, 100, expiry)

	svc := NewSalesService(db, stubSettingsService{}, nil)
	err := svc.ProcessCheckout(requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: product.ID, Quantity: 2, UnitID: &kg.ID},
		},
		PaymentMethod: "cash",
	}, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible")
}
