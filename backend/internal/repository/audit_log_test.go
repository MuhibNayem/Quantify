package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupAuditTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:audit-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.AuditLog{}))
	return db
}

func TestWriteAudit_PersistsRow(t *testing.T) {
	db := setupAuditTestDB(t)
	ctx := context.Background()

	entry := &domain.AuditLog{
		Action:    "DISCOUNT",
		Entity:    "Order",
		EntityID:  "ORD-123",
		UserID:    42,
		Changes:   `{"discount_pct":15.5}`,
		IPAddress: "127.0.0.1",
		UserAgent: "curl/8.0",
		Timestamp: time.Now(),
	}

	require.NoError(t, WriteAudit(ctx, db, entry))

	var stored domain.AuditLog
	require.NoError(t, db.First(&stored, entry.ID).Error)
	assert.Equal(t, "DISCOUNT", stored.Action)
	assert.Equal(t, "Order", stored.Entity)
	assert.Equal(t, "ORD-123", stored.EntityID)
	assert.Equal(t, uint(42), stored.UserID)
	assert.Equal(t, `{"discount_pct":15.5}`, stored.Changes)
	assert.Equal(t, "127.0.0.1", stored.IPAddress)
	assert.Equal(t, "curl/8.0", stored.UserAgent)
	assert.False(t, stored.Timestamp.IsZero())
}

func TestWriteAudit_DBErrorLogsButDoesNotReturn(t *testing.T) {
	// Close the underlying DB to force errors. WriteAudit must swallow failures
	// (best-effort audit) and log a warning — never bubble up to callers.
	hook := test.NewGlobal()
	defer hook.Reset()

	dsn := fmt.Sprintf("file:audit-fail-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.AuditLog{}))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	entry := &domain.AuditLog{
		Action:   "DISCOUNT",
		Entity:   "Order",
		EntityID: "ORD-456",
		UserID:   7,
	}

	// Must NOT return an error.
	require.NoError(t, WriteAudit(context.Background(), db, entry))

	// Must have logged a warning.
	var foundWarn bool
	for _, e := range hook.Entries {
		if e.Level == logrus.WarnLevel {
			foundWarn = true
			break
		}
	}
	assert.True(t, foundWarn, "WriteAudit must log a warning when the DB write fails")
}
