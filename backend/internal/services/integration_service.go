package services

import (
	"inventory/backend/internal/integrations"
	"sync"
)

// IntegrationService manages the lifecycle of third-party integrations.
type IntegrationService struct {
	integrations map[string]integrations.Integration
	mu           sync.RWMutex
}

// NewIntegrationService creates a new IntegrationService.
func NewIntegrationService() *IntegrationService {
	return &IntegrationService{
		integrations: make(map[string]integrations.Integration),
	}
}

// RegisterIntegration registers a new integration.
func (s *IntegrationService) RegisterIntegration(integration integrations.Integration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.integrations[integration.Name()] = integration
}

// GetIntegration returns an integration by name.
func (s *IntegrationService) GetIntegration(name string) (integrations.Integration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	integration, ok := s.integrations[name]
	if !ok {
		return nil, &integrations.ErrIntegrationNotFound{Name: name}
	}
	return integration, nil
}

// HandleWebhook handles an incoming webhook.
func (s *IntegrationService) HandleWebhook(payload *integrations.WebhookPayload) error {
	integration, err := s.GetIntegration(payload.Source)
	if err != nil {
		return err
	}
	return integration.HandleWebhook(payload)
}
