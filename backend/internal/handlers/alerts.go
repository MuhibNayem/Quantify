package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

var (
	allowedAlertTypes = map[string]struct{}{
		"LOW_STOCK":    {},
		"OUT_OF_STOCK": {},
		"OVERSTOCK":    {},
		"EXPIRY_ALERT": {},
	}
	allowedAlertStatuses = map[string]struct{}{
		"ACTIVE":   {},
		"RESOLVED": {},
	}
)

// AlertTriggeredPayload defines the payload for alert triggered events.
type AlertTriggeredPayload struct {
	ProductID uint   `json:"productId"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

// PutProductAlertSettings godoc
// @Summary Configure alert thresholds for a product
// @Description Configures low-stock, overstock, and expiry alert thresholds for a specific product
// @Tags alerts
// @Accept json
// @Produce json
// @Param productId path int true "Product ID"
// @Param settings body requests.ProductAlertSettingsRequest true "Product alert settings"
// @Success 200 {object} domain.ProductAlertSettings
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /products/{productId}/alert-settings [put]
func PutProductAlertSettings(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid product ID", http.StatusBadRequest, err))
		return
	}

	var req requests.ProductAlertSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var product domain.Product
	if err := repository.DB.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Product not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch product", http.StatusInternalServerError, err))
		return
	}

	settings := domain.ProductAlertSettings{
		ProductID:       uint(productID),
		LowStockLevel:   req.LowStockLevel,
		OverStockLevel:  req.OverStockLevel,
		ExpiryAlertDays: req.ExpiryAlertDays,
	}

	// Upsert: create if not exists, update if exists
	if err := repository.DB.Where(domain.ProductAlertSettings{ProductID: uint(productID)}).Assign(settings).FirstOrCreate(&settings).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to save product alert settings", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// ListAlerts godoc
// @Summary Get a list of all active alerts
// @Description Retrieves a list of all active stock-related alerts
// @Tags alerts
// @Accept json
// @Produce json
// @Param type query string false "Filter by alert type (LOW_STOCK, OUT_OF_STOCK, OVERSTOCK, EXPIRY_ALERT)"
// @Param status query string false "Filter by alert status (ACTIVE, RESOLVED)"
// @Param productId query int false "Filter by Product ID"
// @Success 200 {array} domain.Alert
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /alerts [get]
func ListAlerts(c *gin.Context) {
	var alerts []domain.Alert
	db := repository.DB.Preload("Product").Preload("Batch")

	if alertType := c.Query("type"); alertType != "" {
		if _, ok := allowedAlertTypes[alertType]; !ok {
			c.Error(appErrors.NewAppError("Invalid alert type", http.StatusBadRequest, nil))
			return
		}
		db = db.Where("type = ?", alertType)
	}
	if status := c.Query("status"); status != "" {
		if _, ok := allowedAlertStatuses[status]; !ok {
			c.Error(appErrors.NewAppError("Invalid alert status", http.StatusBadRequest, nil))
			return
		}
		db = db.Where("status = ?", status)
	} else {
		db = db.Where("status = ?", "ACTIVE")
	}
	if productID := c.Query("productId"); productID != "" {
		db = db.Where("product_id = ?", productID)
	}

	if err := db.Find(&alerts).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch alerts", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetAlert godoc
// @Summary Get an alert by ID
// @Description Retrieves details of a specific alert by its ID
// @Tags alerts
// @Accept json
// @Produce json
// @Param alertId path int true "Alert ID"
// @Success 200 {object} domain.Alert
// @Failure 404 {object} map[string]interface{} "Alert not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /alerts/{alertId} [get]
func GetAlert(c *gin.Context) {
	alertID := c.Param("alertId")
	var alert domain.Alert
	if err := repository.DB.Preload("Product").Preload("Batch").First(&alert, alertID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Alert not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch alert", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, alert)
}

// ResolveAlert godoc
// @Summary Resolve an alert
// @Description Marks a specific alert as resolved
// @Tags alerts
// @Accept json
// @Produce json
// @Param alertId path int true "Alert ID"
// @Success 200 {object} domain.Alert
// @Failure 404 {object} map[string]interface{} "Alert not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /alerts/{alertId}/resolve [patch]
func ResolveAlert(c *gin.Context) {
	alertID := c.Param("alertId")
	var alert domain.Alert
	if err := repository.DB.First(&alert, alertID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Alert not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch alert", http.StatusInternalServerError, err))
		return
	}

	if alert.Status == "RESOLVED" {
		c.JSON(http.StatusOK, alert) // Already resolved
		return
	}

	if err := repository.DB.Model(&alert).Update("Status", "RESOLVED").Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to resolve alert", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, alert)
}

// PutUserNotificationSettings godoc
// @Summary Configure user notification preferences
// @Description Configures email and SMS notification preferences for a user
// @Tags users
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param settings body requests.UserNotificationSettingsRequest true "User notification settings"
// @Success 200 {object} domain.UserNotificationSettings
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{userId}/notification-settings [put]
func PutUserNotificationSettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid user ID", http.StatusBadRequest, err))
		return
	}

	var req requests.UserNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var user domain.User
	if err := repository.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	settings := domain.UserNotificationSettings{
		UserID:                    uint(userID),
		EmailNotificationsEnabled: req.EmailNotificationsEnabled,
		SMSNotificationsEnabled:   req.SMSNotificationsEnabled,
		EmailAddress:              req.EmailAddress,
		PhoneNumber:               req.PhoneNumber,
	}

	// Upsert: create if not exists, update if exists
	if err := repository.DB.Where(domain.UserNotificationSettings{UserID: uint(userID)}).Assign(settings).FirstOrCreate(&settings).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to save user notification settings", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// CheckAndTriggerAlerts is a background function to check and trigger alerts.
func CheckAndTriggerAlerts() {
	var settings []domain.ProductAlertSettings
	if err := repository.DB.Find(&settings).Error; err != nil {
		logrus.Errorf("Failed to fetch product alert settings: %v", err)
		return
	}

	for _, s := range settings {
		checkAndTriggerAlertsForSettings(&s)
	}
}

// CheckAndTriggerAlertsForProduct evaluates alert thresholds for a single product.
func CheckAndTriggerAlertsForProduct(productID uint) {
	var settings domain.ProductAlertSettings
	if err := repository.DB.Where("product_id = ?", productID).First(&settings).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("Failed to fetch alert settings for product %d: %v", productID, err)
		}
		return
	}

	checkAndTriggerAlertsForSettings(&settings)
}

func checkAndTriggerAlertsForSettings(s *domain.ProductAlertSettings) {
	var product domain.Product
	if err := repository.DB.First(&product, s.ProductID).Error; err != nil {
		logrus.Errorf("Failed to fetch product %d for alert check: %v", s.ProductID, err)
		return
	}

	// Get current quantity
	var currentQuantity int
	repository.DB.Model(&domain.Batch{}).Where("product_id = ?", s.ProductID).Select("sum(quantity)").Row().Scan(&currentQuantity)

	// Low Stock Alert
	if s.LowStockLevel > 0 && currentQuantity <= s.LowStockLevel {
		triggerAlert(s.ProductID, "LOW_STOCK", fmt.Sprintf("Product %s is running low. Current quantity: %d", product.Name, currentQuantity), nil)
	}

	// Out of Stock Alert
	if currentQuantity == 0 {
		triggerAlert(s.ProductID, "OUT_OF_STOCK", fmt.Sprintf("Product %s is out of stock.", product.Name), nil)
	}

	// Overstock Alert
	if s.OverStockLevel > 0 && currentQuantity >= s.OverStockLevel {
		triggerAlert(s.ProductID, "OVERSTOCK", fmt.Sprintf("Product %s is overstocked. Current quantity: %d", product.Name, currentQuantity), nil)
	}

	// Expiry Alert
	if s.ExpiryAlertDays > 0 {
		var expiringBatches []domain.Batch
		expiryThreshold := time.Now().AddDate(0, 0, s.ExpiryAlertDays)
		if err := repository.DB.Where("product_id = ? AND expiry_date IS NOT NULL AND expiry_date <= ?", s.ProductID, expiryThreshold).Find(&expiringBatches).Error; err != nil {
			logrus.Errorf("Failed to fetch expiring batches for product %d: %v", s.ProductID, err)
			return
		}

		for _, batch := range expiringBatches {
			triggerAlert(s.ProductID, "EXPIRY_ALERT", fmt.Sprintf("Batch %s of product %s is expiring soon on %s", batch.BatchNumber, product.Name, batch.ExpiryDate.Format("2006-01-02")), &batch.ID)
		}
	}
}

func triggerAlert(productID uint, alertType, message string, batchID *uint) {
	// Check if an active alert of this type already exists for this product/batch
	var existingAlert domain.Alert
	query := repository.DB.Where("product_id = ? AND type = ? AND status = ?", productID, alertType, "ACTIVE")
	if batchID != nil {
		query = query.Where("batch_id = ?", *batchID)
	}

	if err := query.First(&existingAlert).Error; err == nil {
		// Alert already active, do not re-trigger
		return
	} else if err != gorm.ErrRecordNotFound {
		logrus.Errorf("Database error checking for existing alert: %v", err)
		return
	}

	// Create new alert record
	alert := domain.Alert{
		ProductID:   productID,
		Type:        alertType,
		Message:     message,
		TriggeredAt: time.Now(),
		Status:      "ACTIVE",
		BatchID:     batchID,
	}
	if err := repository.DB.Create(&alert).Error; err != nil {
		logrus.Errorf("Failed to create alert record: %v", err)
		return
	}

	// Publish AlertTriggeredEvent
	payload := AlertTriggeredPayload{
		ProductID: productID,
		Type:      alertType,
		Message:   message,
	}
	if err := message_broker.Publish(context.Background(), "inventory", "alert.triggered", payload); err != nil {
		logrus.Errorf("Failed to publish alert triggered event: %v", err)
	}
	logrus.Infof("Alert triggered and published: %s for product %d", alertType, productID)
}
