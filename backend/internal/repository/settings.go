package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

type SettingsRepository interface {
	GetSettingsByGroup(group string) ([]domain.SystemSetting, error)
	GetAllSettings() ([]domain.SystemSetting, error)
	GetSettingByKey(key string) (*domain.SystemSetting, error)
	UpdateSetting(key, value string, userID uint) error
}

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) GetSettingsByGroup(group string) ([]domain.SystemSetting, error) {
	var settings []domain.SystemSetting
	// Quote "group" to avoid SQL keyword conflict
	err := r.db.Where("\"group\" = ?", group).Find(&settings).Error
	return settings, err
}

func (r *settingsRepository) GetAllSettings() ([]domain.SystemSetting, error) {
	var settings []domain.SystemSetting
	err := r.db.Find(&settings).Error
	return settings, err
}

// UpdateSetting writes a new value for the given setting key. If the key
// does not exist, a new row is created (action="CREATE" audit). If it
// already exists, the value is updated (action="UPDATE" audit).
//
// An audit_log row is written on every successful change. Audit failures
// are logged via logrus but never bubble up — the primary setting write
// must not be blocked by audit problems.
func (r *settingsRepository) UpdateSetting(key, value string, userID uint) error {
	var existing domain.SystemSetting
	err := r.db.Where("key = ?", key).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			setting := domain.SystemSetting{
				Key:   key,
				Value: value,
				Group: "General",
				Type:  "string",
			}
			if createErr := r.db.Create(&setting).Error; createErr != nil {
				return createErr
			}
			r.writeSettingsAudit(context.Background(), "CREATE", key, "", value, userID)
			return nil
		}
		return err
	}
	oldValue := existing.Value
	existing.Value = value
	if saveErr := r.db.Save(&existing).Error; saveErr != nil {
		return saveErr
	}
	r.writeSettingsAudit(context.Background(), "UPDATE", key, oldValue, value, userID)
	return nil
}

func (r *settingsRepository) writeSettingsAudit(ctx context.Context, action, key, oldValue, newValue string, userID uint) {
	changes, err := json.Marshal(map[string]string{
		"old": oldValue,
		"new": newValue,
	})
	if err != nil {
		logrus.WithError(err).WithField("setting_key", key).
			Warn("settings audit: failed to marshal changes")
		return
	}
	entry := &domain.AuditLog{
		Action:    action,
		Entity:    "Settings",
		EntityID:  key,
		UserID:    userID,
		Changes:   string(changes),
		Timestamp: time.Now(),
	}
	_ = WriteAudit(ctx, r.db, entry)
}

func (r *settingsRepository) GetSettingByKey(key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	err := r.db.Where("key = ?", key).First(&setting).Error
	return &setting, err
}
