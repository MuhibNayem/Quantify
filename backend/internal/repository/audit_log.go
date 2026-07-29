package repository

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

// AuditLogFilter holds optional filters and pagination for listing audit logs.
type AuditLogFilter struct {
	Action string
	Entity string
	UserID uint
	Start  time.Time
	End    time.Time
	Page   int
	Limit  int
}

// ListAuditLogs returns paginated, filtered audit logs (newest first).
func ListAuditLogs(ctx context.Context, db *gorm.DB, params AuditLogFilter) ([]domain.AuditLog, int64, error) {
	q := db.WithContext(ctx).Model(&domain.AuditLog{}).Preload("User")
	if params.Action != "" {
		q = q.Where("action = ?", params.Action)
	}
	if params.Entity != "" {
		q = q.Where("entity = ?", params.Entity)
	}
	if params.UserID != 0 {
		q = q.Where("user_id = ?", params.UserID)
	}
	if !params.Start.IsZero() {
		q = q.Where("timestamp >= ?", params.Start)
	}
	if !params.End.IsZero() {
		q = q.Where("timestamp <= ?", params.End)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 200 {
		params.Limit = 50
	}

	var logs []domain.AuditLog
	err := q.Order("timestamp DESC").
		Offset((params.Page - 1) * params.Limit).
		Limit(params.Limit).
		Find(&logs).Error
	return logs, total, err
}

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
