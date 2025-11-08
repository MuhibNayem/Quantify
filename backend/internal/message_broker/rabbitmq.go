package message_broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"inventory/backend/internal/websocket"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

var RabbitMQChannel *amqp091.Channel

func InitRabbitMQ(url string) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		logrus.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	RabbitMQChannel, err = conn.Channel()
	if err != nil {
		logrus.Fatalf("Failed to open a channel: %v", err)
	}

	logrus.Info("Connected to RabbitMQ!")
}

func Publish(exchange, routingKey string, body interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body to json: %w", err)
	}

	return RabbitMQChannel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        jsonBody,
		},
	)
}

func Subscribe(exchange, queueName, routingKey string, hub *websocket.Hub, handler func(d amqp091.Delivery)) error {
	if err := RabbitMQChannel.ExchangeDeclare(
		exchange,
		"topic", // allows dot-delimited routing keys like bulk.import
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
	}

	q, err := RabbitMQChannel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = RabbitMQChannel.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind a queue: %w", err)
	}

	msgs, err := RabbitMQChannel.Consume(
		q.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			handler(d)
			hub.Broadcast(d.Body)
		}
	}()

	return nil
}
