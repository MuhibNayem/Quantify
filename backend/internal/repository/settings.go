package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type SettingsRepository interface {
	GetSettingsByGroup(group string) ([]domain.SystemSetting, error)
	GetAllSettings() ([]domain.SystemSetting, error)
	GetSettingByKey(key string) (*domain.SystemSetting, error)
	UpdateSetting(key, value string) error
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

func (r *settingsRepository) UpdateSetting(key, value string) error {
	var setting domain.SystemSetting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			setting = domain.SystemSetting{
				Key:   key,
				Value: value,
				Group: "General",
				Type:  "string",
			}
			return r.db.Create(&setting).Error
		}
		return err
	}
	setting.Value = value
	return r.db.Save(&setting).Error
}

func (r *settingsRepository) GetSettingByKey(key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	err := r.db.Where("key = ?", key).First(&setting).Error
	return &setting, err
}
