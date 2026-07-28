package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	appErrors "inventory/backend/internal/errors"
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
		c.Error(appErrors.NewAppError("Failed to fetch settings", http.StatusInternalServerError, err))
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
		c.Error(appErrors.NewAppError("Invalid payload", http.StatusBadRequest, err))
		return
	}
	descriptor, ok := settingDescriptors[req.Key]
	if !ok {
		c.Error(appErrors.NewAppError("Unknown setting: "+req.Key, http.StatusBadRequest, nil))
		return
	}
	valueStr := fmt.Sprintf("%v", req.Value)
	if descriptor.Validate != nil {
		if err := descriptor.Validate(valueStr); err != nil {
			c.Error(appErrors.NewAppError("Invalid value: "+err.Error(), http.StatusBadRequest, err))
			return
		}
	}
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	userID := authUserID.(uint)
	if err := h.service.UpdateSetting(req.Key, valueStr, userID); err != nil {
		c.Error(appErrors.NewAppError("Failed to update setting", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Setting updated"})
}

func (h *SettingsHandler) GetPublicConfigurations(c *gin.Context) {
	settings, err := h.service.GetPublicSettings()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch configurations", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, settings)
}
