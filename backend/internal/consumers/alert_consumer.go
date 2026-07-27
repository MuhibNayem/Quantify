package consumers

import (
	"context"
	"encoding/json"
	"time"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/notifications"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/websocket"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

// loadSubscribedUsers returns the list of users subscribed to the given alert type,
// resolved by joining AlertRoleSubscription -> Role -> User via users.role_id,
// and excluding any user who has explicitly opted out via
// UserNotificationSubscription(Enabled = false).
func loadSubscribedUsers(ctx context.Context, db *gorm.DB, alertType string) ([]domain.User, error) {
	var subscriptions []domain.AlertRoleSubscription
	if err := db.WithContext(ctx).Where("alert_type = ?", alertType).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}

	roleNames := make([]string, 0, len(subscriptions))
	for _, s := range subscriptions {
		roleNames = append(roleNames, s.Role)
	}

	var users []domain.User
	err := db.WithContext(ctx).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", roleNames).
		Where("NOT EXISTS (SELECT 1 FROM user_notification_subscriptions s WHERE s.user_id = users.id AND s.type = ? AND s.enabled = false)", alertType).
		Find(&users).Error
	return users, err
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

	users, err := loadSubscribedUsers(context.Background(), c.DB, payload.Type)
	if err != nil {
		logrus.Errorf("Failed to load subscribed users for alert type %s: %v", payload.Type, err)
		d.Nack(true, true)
		return
	}

	if len(users) == 0 {
		logrus.Infof("No subscribed users found for alert type %s", payload.Type)
		d.Ack(false) // No one is subscribed, so we're done.
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
