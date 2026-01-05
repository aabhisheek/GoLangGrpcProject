package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config holds RabbitMQ configuration
type Config struct {
	Host          string
	Port          int
	User          string
	Password      string
	Exchange      string
	Queue         string
	DLQ           string
	PrefetchCount int
}

// RabbitMQ wraps amqp.Connection and provides messaging functionality
type RabbitMQ struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	config        Config
	reconnectChan chan struct{}
}

// NewRabbitMQ creates a new RabbitMQ client
func NewRabbitMQ(cfg Config) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	mq := &RabbitMQ{
		conn:          conn,
		channel:       ch,
		config:        cfg,
		reconnectChan: make(chan struct{}),
	}

	// Setup exchange and queues
	if err := mq.setup(); err != nil {
		mq.Close()
		return nil, fmt.Errorf("failed to setup RabbitMQ: %w", err)
	}

	return mq, nil
}

// setup declares exchange, queues, and bindings
func (mq *RabbitMQ) setup() error {
	// Declare exchange
	err := mq.channel.ExchangeDeclare(
		mq.config.Exchange, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare DLQ
	_, err = mq.channel.QueueDeclare(
		mq.config.DLQ, // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Declare main queue with DLQ
	args := amqp.Table{
		"x-dead-letter-exchange":    mq.config.Exchange,
		"x-dead-letter-routing-key": mq.config.DLQ,
	}

	_, err = mq.channel.QueueDeclare(
		mq.config.Queue, // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		args,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = mq.channel.QueueBind(
		mq.config.Queue,    // queue name
		mq.config.Queue,    // routing key
		mq.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Bind DLQ
	err = mq.channel.QueueBind(
		mq.config.DLQ,      // queue name
		mq.config.DLQ,      // routing key
		mq.config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// Set QoS
	err = mq.channel.Qos(
		mq.config.PrefetchCount, // prefetch count
		0,                       // prefetch size
		false,                   // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (mq *RabbitMQ) Close() error {
	if mq.channel != nil {
		mq.channel.Close()
	}
	if mq.conn != nil {
		return mq.conn.Close()
	}
	return nil
}

// Publish publishes a message to the queue
func (mq *RabbitMQ) Publish(ctx context.Context, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = mq.channel.PublishWithContext(
		ctx,
		mq.config.Exchange, // exchange
		routingKey,         // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume consumes messages from the queue
func (mq *RabbitMQ) Consume(ctx context.Context, handler MessageHandler) error {
	msgs, err := mq.channel.Consume(
		mq.config.Queue, // queue
		"",              // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			// Process message with handler
			err := handler(ctx, msg.Body)
			if err != nil {
				// Reject and requeue on error
				msg.Nack(false, true)
				continue
			}

			// Acknowledge successful processing
			msg.Ack(false)
		}
	}
}

// ConsumeWithRetry consumes messages with retry logic
func (mq *RabbitMQ) ConsumeWithRetry(ctx context.Context, handler MessageHandler, maxRetries int) error {
	msgs, err := mq.channel.Consume(
		mq.config.Queue, // queue
		"",              // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			// Get retry count from headers
			retryCount := 0
			if msg.Headers != nil {
				if count, ok := msg.Headers["x-retry-count"].(int32); ok {
					retryCount = int(count)
				}
			}

			// Process message with handler
			err := handler(ctx, msg.Body)
			if err != nil {
				if retryCount < maxRetries {
					// Retry: reject and requeue with incremented count
					retryCount++
					headers := amqp.Table{
						"x-retry-count": int32(retryCount),
					}
					
					// Republish with updated headers
					mq.channel.PublishWithContext(
						ctx,
						mq.config.Exchange,
						mq.config.Queue,
						false,
						false,
						amqp.Publishing{
							ContentType:  msg.ContentType,
							Body:         msg.Body,
							DeliveryMode: msg.DeliveryMode,
							Headers:      headers,
						},
					)
					
					msg.Ack(false)
				} else {
					// Send to DLQ after max retries
					msg.Nack(false, false)
				}
				continue
			}

			// Acknowledge successful processing
			msg.Ack(false)
		}
	}
}

// MessageHandler is a function that processes a message
type MessageHandler func(ctx context.Context, body []byte) error

// IsConnected checks if the connection is alive
func (mq *RabbitMQ) IsConnected() bool {
	return mq.conn != nil && !mq.conn.IsClosed()
}

