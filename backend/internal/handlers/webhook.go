package handlers

import (
	"inventory/backend/internal/integrations"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles incoming webhooks from third-party integrations.
type WebhookHandler struct {
	integrationService *services.IntegrationService
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(integrationService *services.IntegrationService) *WebhookHandler {
	return &WebhookHandler{integrationService: integrationService}
}

// HandleWebhook handles an incoming webhook.
func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	var payload integrations.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(appErrors.NewAppError("Invalid webhook payload", http.StatusBadRequest, err))
		return
	}

	if err := h.integrationService.HandleWebhook(&payload); err != nil {
		c.Error(appErrors.NewAppError("Failed to handle webhook", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook received"})
}
