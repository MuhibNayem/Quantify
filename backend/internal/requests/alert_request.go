package requests

// ProductAlertSettingsRequest represents the request body for configuring product alert thresholds.
type ProductAlertSettingsRequest struct {
	LowStockLevel   int `json:"lowStockLevel"`
	OverStockLevel  int `json:"overStockLevel"`
	ExpiryAlertDays int `json:"expiryAlertDays"`
}

// UserNotificationSettingsRequest represents the request body for configuring user notification preferences.
type UserNotificationSettingsRequest struct {
	EmailNotificationsEnabled bool   `json:"emailNotificationsEnabled"`
	SMSNotificationsEnabled   bool   `json:"smsNotificationsEnabled"`
	EmailAddress              string `json:"emailAddress"`
	PhoneNumber               string `json:"phoneNumber"`
}
