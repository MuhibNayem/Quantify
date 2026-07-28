package handlers

import (
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

func setupAlertTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:alerts-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Product{},
		&domain.Batch{},
		&domain.Alert{},
		&domain.ProductAlertSettings{},
	))

	previousDB := repository.DB()
	repository.SetDB(db)
	t.Cleanup(func() {
		repository.SetDB(previousDB)
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			require.NoError(t, sqlDB.Close())
		}
	})

	return db
}

func seedAlertTestProduct(t *testing.T, db *gorm.DB, sku string) domain.Product {
	t.Helper()
	product := domain.Product{Name: sku, SKU: sku, Status: "Active"}
	require.NoError(t, db.Create(&product).Error)
	return product
}

func seedAlertTestSettings(t *testing.T, db *gorm.DB, productID uint) {
	t.Helper()
	require.NoError(t, db.Create(&domain.ProductAlertSettings{ProductID: productID}).Error)
}

func TestCheckAndTriggerAlerts_ContinuesOnError(t *testing.T) {
	db := setupAlertTestDB(t)
	badProduct := seedAlertTestProduct(t, db, "BAD")
	goodProduct := seedAlertTestProduct(t, db, "GOOD")
	seedAlertTestSettings(t, db, badProduct.ID)
	seedAlertTestSettings(t, db, goodProduct.ID)

	original := alertEvalFn
	alertEvalFn = func(db *gorm.DB, settings domain.ProductAlertSettings) {
		if settings.ProductID == badProduct.ID {
			panic("boom")
		}
		require.NoError(t, db.Create(&domain.Alert{
			ProductID:   settings.ProductID,
			Type:        "OUT_OF_STOCK",
			Message:     "test alert",
			TriggeredAt: time.Now(),
			Status:      "ACTIVE",
		}).Error)
	}
	t.Cleanup(func() { alertEvalFn = original })

	assert.NotPanics(t, CheckAndTriggerAlerts)
	var count int64
	require.NoError(t, db.Model(&domain.Alert{}).Where("product_id = ?", goodProduct.ID).Count(&count).Error)
	assert.Greater(t, count, int64(0))
}
