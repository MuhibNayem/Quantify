package requests

// CreateAlertRoleSubscriptionRequest represents the request body for creating an alert role subscription.
type CreateAlertRoleSubscriptionRequest struct {
	AlertType string `json:"alertType" binding:"required"`
	Role      string `json:"role" binding:"required"`
}
