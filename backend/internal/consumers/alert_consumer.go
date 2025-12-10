package consumers

import (
	"context"
	"encoding/json"
	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/notifications"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/websocket"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"inventory/backend/internal/handlers"
	"inventory/backend/internal/message_broker"
)

type AlertConsumer struct {
	DB               *gorm.DB
	NotificationRepo repository.NotificationRepository
	Hub              *websocket.Hub
	Cfg              *config.Config
}

func NewAlertConsumer(db *gorm.DB, notificationRepo repository.NotificationRepository, hub *websocket.Hub, cfg *config.Config) *AlertConsumer {
	return &AlertConsumer{
		DB:               db,
		NotificationRepo: notificationRepo,
		Hub:              hub,
		Cfg:              cfg,
	}
}

func (c *AlertConsumer) Start(ctx context.Context) context.CancelFunc {
	cancelStock := message_broker.Subscribe(ctx, "inventory", "stock-alerts-queue", "stock.adjusted", c.handleStockAdjusted)
	cancelAlert := message_broker.Subscribe(ctx, "inventory", "alerts", "alert.triggered", c.handleAlertDelivery)

	return func() {
		cancelStock()
		cancelAlert()
	}
}

func (c *AlertConsumer) handleStockAdjusted(ctx context.Context, deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.processStockAdjustedDelivery(d)
		}
	}
}

func (c *AlertConsumer) processStockAdjustedDelivery(d amqp091.Delivery) {
	var payload handlers.StockAdjustedEventPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		logrus.Errorf("AlertConsumer: failed to unmarshal stock adjustment payload: %v", err)
		_ = d.Nack(false, false)
		return
	}

	if payload.ProductID == 0 {
		logrus.Warn("AlertConsumer: received stock adjustment payload without product ID")
		_ = d.Ack(false)
		return
	}

	handlers.CheckAndTriggerAlertsForProduct(payload.ProductID)
	_ = d.Ack(false)
}

func (c *AlertConsumer) handleAlertDelivery(ctx context.Context, deliveries <-chan amqp091.Delivery) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}
			c.processAlertDelivery(d)
		}
	}
}

func (c *AlertConsumer) processAlertDelivery(d amqp091.Delivery) {
	var payload handlers.AlertTriggeredPayload
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		logrus.Errorf("Failed to unmarshal alert payload: %v", err)
		d.Nack(false, false)
		return
	}

	logrus.Infof("Handling alert delivery for alert type %s", payload.Type)

	// 1. Find roles subscribed to this alert type
	var subscriptions []domain.AlertRoleSubscription
	if err := c.DB.Where("alert_type = ?", payload.Type).Find(&subscriptions).Error; err != nil {
		logrus.Errorf("Failed to fetch alert subscriptions: %v", err)
		d.Nack(true, true) // Nack and requeue
		return
	}

	if len(subscriptions) == 0 {
		logrus.Infof("No subscriptions found for alert type %s", payload.Type)
		d.Ack(false) // No one is subscribed, so we're done.
		return
	}

	var roles []string
	for _, sub := range subscriptions {
		roles = append(roles, sub.Role)
	}

	// 2. Find users with those roles
	var users []domain.User
	if err := c.DB.Where("role IN (?)", roles).Find(&users).Error; err != nil {
		logrus.Errorf("Failed to fetch users for roles %v: %v", roles, err)
		d.Nack(true, true)
		return
	}

	// 3. Send notifications to the targeted users
	for _, user := range users {
		var settings domain.UserNotificationSettings
		if err := c.DB.Where("user_id = ?", user.ID).First(&settings).Error; err != nil {
			// If settings don't exist, we can't send notifications, so we skip.
			continue
		}

		// Send email if enabled
		if settings.EmailNotificationsEnabled && settings.EmailAddress != "" {
			subject := "Inventory Alert: " + payload.Type
			body := payload.Message
			if err := notifications.SendEmail(c.Cfg, settings.EmailAddress, subject, body); err != nil {
				logrus.Errorf("Failed to send email to %s: %v", settings.EmailAddress, err)
			}
		}

		// Create and persist in-app notification
		notificationPayload, _ := json.Marshal(payload)
		notification := domain.Notification{
			UserID:      user.ID,
			Type:        "ALERT",
			Title:       "New Alert: " + payload.Type,
			Message:     payload.Message,
			Payload:     string(notificationPayload),
			TriggeredAt: time.Now(),
		}
		if err := c.NotificationRepo.CreateNotification(&notification); err != nil {
			logrus.Errorf("Failed to create in-app notification for user %d: %v", user.ID, err)
		} else {
			// Send user-specific WebSocket message
			c.Hub.SendToUser(user.ID, notification)
		}
	}

	d.Ack(false)
}
