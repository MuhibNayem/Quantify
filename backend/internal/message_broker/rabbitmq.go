package message_broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

var (
	rabbitMQManager *RabbitMQManager
	once            sync.Once
)

// RabbitMQManager handles RabbitMQ connection and channel management.
type RabbitMQManager struct {
	url        string
	conn       *amqp091.Connection
	channel    *amqp091.Channel
	mu         sync.Mutex
	isReady    chan bool
	isClosed   bool
	notifyConn chan *amqp091.Error
	notifyChan chan *amqp091.Error
	cancelFunc context.CancelFunc
}

// NewRabbitMQManager creates a new RabbitMQManager.
func NewRabbitMQManager(url string) *RabbitMQManager {
	m := &RabbitMQManager{
		url:     url,
		isReady: make(chan bool),
	}
	go m.handleReconnect()
	return m
}

// handleReconnect handles the connection and reconnection logic.
func (m *RabbitMQManager) handleReconnect() {
	for {
		m.mu.Lock()
		if m.isClosed {
			m.mu.Unlock()
			return
		}
		m.mu.Unlock()

		logrus.Info("Attempting to connect to RabbitMQ...")
		if err := m.connect(); err != nil {
			logrus.Errorf("Failed to connect to RabbitMQ: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		logrus.Info("Connected to RabbitMQ!")
		close(m.isReady) // Signal that the initial connection is ready

		select {
		case err := <-m.notifyConn:
			logrus.Errorf("RabbitMQ connection lost: %v. Reconnecting...", err)
		case err := <-m.notifyChan:
			logrus.Errorf("RabbitMQ channel lost: %v. Reconnecting...", err)
		}

		m.mu.Lock()
		m.isReady = make(chan bool) // Reset the ready channel for the next connection
		m.mu.Unlock()
	}
}

// connect establishes a connection and channel.
func (m *RabbitMQManager) connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := amqp091.Dial(m.url)
	if err != nil {
		return err
	}
	m.conn = conn
	m.notifyConn = make(chan *amqp091.Error)
	m.conn.NotifyClose(m.notifyConn)

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	m.channel = ch
	m.notifyChan = make(chan *amqp091.Error)
	m.channel.NotifyClose(m.notifyChan)

	return nil
}

// Close gracefully closes the connection and channel.
func (m *RabbitMQManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isClosed = true
	if m.channel != nil {
		m.channel.Close()
	}
	if m.conn != nil {
		m.conn.Close()
	}
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

// InitRabbitMQ initializes the RabbitMQ manager.
func InitRabbitMQ(url string) {
	once.Do(func() {
		rabbitMQManager = NewRabbitMQManager(url)
	})
}

// Close shuts down the RabbitMQ manager if initialized.
func Close() {
	if rabbitMQManager != nil {
		rabbitMQManager.Close()
	}
}

// Publish publishes a message to RabbitMQ.
func Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	<-rabbitMQManager.isReady // Wait for the connection to be ready

	rabbitMQManager.mu.Lock()
	defer rabbitMQManager.mu.Unlock()

	if rabbitMQManager.channel == nil {
		return fmt.Errorf("RabbitMQ channel is not available")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body to json: %w", err)
	}

	return rabbitMQManager.channel.PublishWithContext(ctx,
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

// Subscribe subscribes to a RabbitMQ queue.
// Subscribe subscribes to a RabbitMQ queue and handles reconnections.
func Subscribe(ctx context.Context, exchange, queueName, routingKey string, handler func(context.Context, <-chan amqp091.Delivery)) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			<-rabbitMQManager.isReady // Wait for the connection to be ready

			rabbitMQManager.mu.Lock()
			if rabbitMQManager.channel == nil {
				rabbitMQManager.mu.Unlock()
				time.Sleep(1 * time.Second) // Wait before retrying
				continue
			}

			if err := rabbitMQManager.channel.ExchangeDeclare(
				exchange,
				"topic",
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				logrus.Errorf("Failed to declare exchange %s: %v", exchange, err)
				rabbitMQManager.mu.Unlock()
				continue
			}

			q, err := rabbitMQManager.channel.QueueDeclare(
				queueName,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				logrus.Errorf("Failed to declare a queue: %v", err)
				rabbitMQManager.mu.Unlock()
				continue
			}

			err = rabbitMQManager.channel.QueueBind(
				q.Name,
				routingKey,
				exchange,
				false,
				nil,
			)
			if err != nil {
				logrus.Errorf("Failed to bind a queue: %v", err)
				rabbitMQManager.mu.Unlock()
				continue
			}

			msgs, err := rabbitMQManager.channel.Consume(
				q.Name,
				"",
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				logrus.Errorf("Failed to register a consumer: %v", err)
				rabbitMQManager.mu.Unlock()
				continue
			}
			rabbitMQManager.mu.Unlock()

			logrus.Infof("Subscribed to queue '%s'", queueName)

			handler(ctx, msgs)

			select {
			case <-ctx.Done():
				return
			default:
			}

			time.Sleep(5 * time.Second) // Wait before attempting to re-subscribe
		}
	}()
	return cancel
}
