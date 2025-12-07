package handlers

import (
	"inventory/backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service services.SettingsService
}

func NewSettingsHandler(service services.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	if err := h.service.UpdateSetting(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Setting updated"})
}
