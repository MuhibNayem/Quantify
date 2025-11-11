package handlers

import (
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateAlertRoleSubscription godoc
// @Summary Create an alert role subscription
// @Description Creates a new subscription to link a role with an alert type.
// @Tags alerts
// @Accept json
// @Produce json
// @Param subscription body requests.CreateAlertRoleSubscriptionRequest true "Subscription details"
// @Success 201 {object} domain.AlertRoleSubscription
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 409 {object} map[string]interface{} "Conflict"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /alerts/subscriptions [post]
func CreateAlertRoleSubscription(c *gin.Context) {
	var req requests.CreateAlertRoleSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	subscription := domain.AlertRoleSubscription{
		AlertType: req.AlertType,
		Role:      req.Role,
	}

	if err := repository.DB.Create(&subscription).Error; err != nil {
		if err.Error() == "UNIQUE constraint failed: alert_role_subscriptions.alert_type, alert_role_subscriptions.role" {
			c.Error(appErrors.NewAppError("Subscription already exists", http.StatusConflict, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to create subscription", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// ListAlertRoleSubscriptions godoc
// @Summary List all alert role subscriptions
// @Description Retrieves a list of all current alert-role subscriptions.
// @Tags alerts
// @Produce json
// @Success 200 {array} domain.AlertRoleSubscription
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /alerts/subscriptions [get]
func ListAlertRoleSubscriptions(c *gin.Context) {
	var subscriptions []domain.AlertRoleSubscription
	if err := repository.DB.Find(&subscriptions).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch subscriptions", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// DeleteAlertRoleSubscription godoc
// @Summary Delete an alert role subscription
// @Description Deletes a subscription by its ID.
// @Tags alerts
// @Param id path int true "Subscription ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /alerts/subscriptions/{id} [delete]
func DeleteAlertRoleSubscription(c *gin.Context) {
	id := c.Param("id")
	var subscription domain.AlertRoleSubscription
	if err := repository.DB.First(&subscription, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Subscription not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch subscription", http.StatusInternalServerError, err))
		return
	}

	if err := repository.DB.Delete(&subscription).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete subscription", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}
