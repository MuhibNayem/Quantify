package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"inventory/backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service services.SettingsService
}

func NewSettingsHandler(service services.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

type SettingDescriptor struct {
	Type     string
	Validate func(v string) error
}

var settingDescriptors = map[string]SettingDescriptor{
	"tax_rate_percentage": {
		Type: "number",
		Validate: numericRange(0, 100),
	},
	"currency": {
		Type: "string",
		Validate: func(v string) error {
			if len(v) != 3 {
				return errors.New("must be a 3-letter code")
			}
			return nil
		},
	},
	"currency_code": {
		Type: "string",
		Validate: func(v string) error {
			if len(v) != 3 {
				return errors.New("must be a 3-letter code")
			}
			return nil
		},
	},
	"currency_symbol":           {Type: "string"},
	"business_name":             {Type: "string"},
	"timezone":                  {Type: "string"},
	"locale":                    {Type: "string"},
	"loyalty.tier_silver":       {Type: "number", Validate: numericRange(0, 1000000)},
	"loyalty.tier_gold":         {Type: "number", Validate: numericRange(0, 1000000)},
	"loyalty.tier_platinum":     {Type: "number", Validate: numericRange(0, 1000000)},
	"loyalty.earning_rate":      {Type: "number", Validate: numericRange(0, 1)},
	"loyalty.redemption_rate":   {Type: "number", Validate: numericRange(0, 1)},
	"loyalty.points_earning_rate":    {Type: "number", Validate: numericRange(0, 1)},
	"loyalty.points_redemption_rate": {Type: "number", Validate: numericRange(0, 1)},
	"return_window_days":        {Type: "number", Validate: numericRange(0, 365)},
}

func numericRange(min, max float64) func(string) error {
	return func(v string) error {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return errors.New("not a number")
		}
		if f < min || f > max {
			return fmt.Errorf("must be between %v and %v", min, max)
		}
		return nil
	}
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
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	descriptor, ok := settingDescriptors[req.Key]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown setting: " + req.Key})
		return
	}
	valueStr := fmt.Sprintf("%v", req.Value)
	if descriptor.Validate != nil {
		if err := descriptor.Validate(valueStr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value: " + err.Error()})
			return
		}
	}
	if err := h.service.UpdateSetting(req.Key, valueStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Setting updated"})
}

func (h *SettingsHandler) GetPublicConfigurations(c *gin.Context) {
	settings, err := h.service.GetPublicSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
		return
	}
	c.JSON(http.StatusOK, settings)
}
