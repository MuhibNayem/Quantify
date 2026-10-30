package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupBarcodeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:barcode-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Product{},
		&domain.ProductUnit{},
		&domain.Unit{},
		&domain.UnitDimension{},
	))
	return db
}

func TestResolveBarcode_ProductSKUAndBarcode(t *testing.T) {
	db := setupBarcodeTestDB(t)
	product := domain.Product{
		SKU:        "WIRE-01",
		Name:       "Copper wire",
		BarcodeUPC: "0123456789012",
	}
	require.NoError(t, db.Create(&product).Error)

	repo := NewBarcodeRepository(db)

	bySKU, unit, err := repo.ResolveBarcode("WIRE-01")
	require.NoError(t, err)
	assert.Equal(t, product.ID, bySKU.ID)
	assert.Nil(t, unit)

	byBarcode, unit, err := repo.ResolveBarcode("0123456789012")
	require.NoError(t, err)
	assert.Equal(t, product.ID, byBarcode.ID)
	assert.Nil(t, unit)
}

func TestResolveBarcode_SellingUnitBarcode(t *testing.T) {
	db := setupBarcodeTestDB(t)
	product := domain.Product{SKU: "WIRE-02", Name: "Copper wire"}
	require.NoError(t, db.Create(&product).Error)

	dimension := domain.UnitDimension{Code: "length", Name: "Length"}
	require.NoError(t, db.Create(&dimension).Error)
	unit := domain.Unit{Code: "in", Name: "Inch", Symbol: "in", DimensionID: dimension.ID, Factor: 1.0 / 12.0}
	require.NoError(t, db.Create(&unit).Error)

	productUnit := domain.ProductUnit{
		ProductID: product.ID,
		UnitID:    unit.ID,
		Factor:    1.0 / 12.0,
		Price:     0.5,
		Barcode:   "7777777777777",
		IsActive:  true,
	}
	require.NoError(t, db.Create(&productUnit).Error)

	repo := NewBarcodeRepository(db)
	resolvedProduct, resolvedUnit, err := repo.ResolveBarcode("7777777777777")
	require.NoError(t, err)
	assert.Equal(t, product.ID, resolvedProduct.ID)
	require.NotNil(t, resolvedUnit)
	assert.Equal(t, productUnit.ID, resolvedUnit.ID)
	assert.Equal(t, unit.ID, resolvedUnit.UnitID)
	assert.Equal(t, "in", resolvedUnit.Unit.Code)

	// Inactive unit barcodes must not resolve.
	require.NoError(t, db.Model(&domain.ProductUnit{}).
		Where("id = ?", productUnit.ID).Update("is_active", false).Error)
	_, _, err = repo.ResolveBarcode("7777777777777")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestResolveBarcode_Unknown(t *testing.T) {
	db := setupBarcodeTestDB(t)
	repo := NewBarcodeRepository(db)
	_, _, err := repo.ResolveBarcode("does-not-exist")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
