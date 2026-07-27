package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.SystemSetting{},
		&domain.AuditLog{},
	))
	return db
}

func TestUpdateSetting_WritesAuditRowForUpdate(t *testing.T) {
	db := setupSettingsTestDB(t)
	repo := NewSettingsRepository(db)

	require.NoError(t, db.Create(&domain.SystemSetting{
		Key:   "tax_rate_percentage",
		Value: "5",
		Group: "Tax",
		Type:  "number",
	}).Error)

	require.NoError(t, repo.UpdateSetting("tax_rate_percentage", "10.5", 42))

	var logs []domain.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1, "exactly one audit row expected for an update")
	assert.Equal(t, "UPDATE", logs[0].Action)
	assert.Equal(t, "Settings", logs[0].Entity)
	assert.Equal(t, "tax_rate_percentage", logs[0].EntityID)
	assert.Equal(t, uint(42), logs[0].UserID)
	assert.Contains(t, logs[0].Changes, `"old":"5"`, "old value should be in changes JSON")
	assert.Contains(t, logs[0].Changes, `"new":"10.5"`, "new value should be in changes JSON")
}

func TestUpdateSetting_WritesAuditRowForCreate(t *testing.T) {
	db := setupSettingsTestDB(t)
	repo := NewSettingsRepository(db)

	require.NoError(t, repo.UpdateSetting("currency_symbol", "৳", 7))

	var logs []domain.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1, "audit row expected when a setting is first created")
	assert.Equal(t, "CREATE", logs[0].Action)
	assert.Equal(t, "Settings", logs[0].Entity)
	assert.Equal(t, "currency_symbol", logs[0].EntityID)
	assert.Equal(t, uint(7), logs[0].UserID)
}
