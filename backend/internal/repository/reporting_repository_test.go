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

func setupReportingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:reporting-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Supplier{},
		&domain.Product{},
		&domain.Batch{},
		&domain.StockAdjustment{},
	))
	return db
}

func TestProductPerformanceReport_Succeeds(t *testing.T) {
	db := setupReportingTestDB(t)

	supplier := domain.Supplier{Name: "Acme"}
	require.NoError(t, db.Create(&supplier).Error)
	product := domain.Product{
		SKU:           "SKU-1",
		Name:          "Widget",
		SupplierID:    supplier.ID,
		PurchasePrice: 5,
		SellingPrice:  10,
	}
	require.NoError(t, db.Create(&product).Error)

	repo := NewReportsRepository(db)
	start := time.Now().Add(-30 * 24 * time.Hour)
	end := time.Now()
	report, err := repo.GetProductPerformanceAnalytics(start, end, "", 0)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotEmpty(t, report, "expected at least the seeded product in the report")
}