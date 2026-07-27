package repository

import (
	"context"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

// WriteAudit persists an audit log entry.
//
// Best-effort: the function is designed for sensitive action call-sites where
// the primary operation must not be blocked or rolled back by audit failure.
// On any error, the failure is logged via logrus.WithError(...).Warn and the
// function returns nil so the caller can continue.
//
// Callers should still populate as many fields as they have available
// (Entity, EntityID, UserID, Changes, IPAddress, UserAgent). The Timestamp
// field will be set to time.Now() if it is zero.
func WriteAudit(ctx context.Context, db *gorm.DB, entry *domain.AuditLog) error {
	if entry == nil {
		return nil
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = nowFn()
	}
	if err := db.WithContext(ctx).Create(entry).Error; err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"action":    entry.Action,
			"entity":    entry.Entity,
			"entity_id": entry.EntityID,
			"user_id":   entry.UserID,
		}).Warn("audit log write failed")
		return nil
	}
	return nil
}
