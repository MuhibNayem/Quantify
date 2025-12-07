package services

import (
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

type SettingsService interface {
	GetAllSettings() (map[string][]domain.SystemSetting, error)
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
