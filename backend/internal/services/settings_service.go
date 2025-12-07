package services

import (
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

type SettingsService interface {
	GetAllSettings() (map[string][]domain.SystemSetting, error)
	GetSetting(key string) (string, error)
	GetPublicSettings() (map[string]string, error)
	UpdateSetting(key, value string) error
}

type settingsService struct {
	repo repository.SettingsRepository
}

func NewSettingsService(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

func (s *settingsService) GetAllSettings() (map[string][]domain.SystemSetting, error) {
	settings, err := s.repo.GetAllSettings()
	if err != nil {
		return nil, err
	}
	// Group settings
	grouped := make(map[string][]domain.SystemSetting)
	for _, setting := range settings {
		grouped[setting.Group] = append(grouped[setting.Group], setting)
	}
	return grouped, nil
}

func (s *settingsService) UpdateSetting(key, value string) error {
	return s.repo.UpdateSetting(key, value)
}

func (s *settingsService) GetSetting(key string) (string, error) {
	setting, err := s.repo.GetSettingByKey(key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *settingsService) GetPublicSettings() (map[string]string, error) {
	settings, err := s.repo.GetAllSettings()
	if err != nil {
		return nil, err
	}

	publicKeys := map[string]bool{
		"currency_symbol":                true,
		"timezone":                       true,
		"locale":                         true,
		"business_name":                  true,
		"return_window_days":             true,
		"tax_rate_percentage":            true,
		"loyalty_points_earning_rate":    true,
		"loyalty_points_redemption_rate": true,
		"loyalty_tier_silver":            true,
		"loyalty_tier_gold":              true,
		"loyalty_tier_platinum":          true,
	}

	publicSettings := make(map[string]string)
	for _, setting := range settings {
		if publicKeys[setting.Key] {
			publicSettings[setting.Key] = setting.Value
		}
	}
	return publicSettings, nil
}
