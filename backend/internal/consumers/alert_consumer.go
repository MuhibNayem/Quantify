package consumers

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"inventory/backend/internal/handlers"
	"inventory/backend/internal/message_broker"
)

type AlertConsumer struct{}

func NewAlertConsumer() *AlertConsumer {
	return &AlertConsumer{}
}

func (c *AlertConsumer) Start(ctx context.Context) context.CancelFunc {
	return message_broker.Subscribe(ctx, "inventory", "stock-alerts-queue", "stock.adjusted", c.handleStockAdjusted)
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
