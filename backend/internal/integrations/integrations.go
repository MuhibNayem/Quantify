package integrations

// Integration is the interface that all third-party integrations must implement.
type Integration interface {
	// Name returns the name of the integration.
	Name() string
	// Connect initializes the integration with the provided settings.
	Connect(settings map[string]string) error
	// Disconnect cleans up the integration.
	Disconnect() error
	// SyncProducts syncs products with the third-party service.
	SyncProducts() error
	// SyncOrders syncs orders with the third-party service.
	SyncOrders() error
	// HandleWebhook handles incoming webhooks from the third-party service.
	HandleWebhook(payload *WebhookPayload) error
}

// WebhookPayload represents the data received from a webhook.
type WebhookPayload struct {
	// Source is the name of the integration that sent the webhook.
	Source string
	// Event is the type of event that triggered the webhook.
	Event string
	// Data is the payload of the webhook.
	Data map[string]interface{}
}

// NewIntegration creates a new integration based on the provided name.
func NewIntegration(name string) (Integration, error) {
	switch name {
	// Add cases for each supported integration here.
	// For example:
	// case "shopify":
	// 	return &ShopifyIntegration{}, nil
	default:
		return nil, &ErrIntegrationNotFound{Name: name}
	}
}

// ErrIntegrationNotFound is returned when an integration is not found.
type ErrIntegrationNotFound struct {
	Name string
}

func (e *ErrIntegrationNotFound) Error() string {
	return "integration not found: " + e.Name
}
