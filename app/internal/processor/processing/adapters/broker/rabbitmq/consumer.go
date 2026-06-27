package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	easyjson "github.com/mailru/easyjson"
	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/converter/generated"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

const (
	exchangeType           = "topic"
	deadLetterExchangeType = "direct"
)

var (
	errBrokerNotConfigured = errors.New("broker is not configured")
	errBrokerNotConnected  = errors.New("broker is not connected")
)

var _ appport.EventConsumer = (*Consumer)(nil)

// Consumer reads avatar lifecycle events from RabbitMQ.
type Consumer struct {
	cfg  Config
	log  appport.Logger
	conv converter.MessageConverter

	conn       *amqp091.Connection
	ch         *amqp091.Channel
	deliveries <-chan amqp091.Delivery
}

// NewConsumer creates a RabbitMQ event consumer.
func NewConsumer(cfg Config, log appport.Logger) (*Consumer, error) {
	c := &Consumer{
		cfg:  cfg,
		log:  log,
		conv: &generated.MessageConverterImpl{},
	}
	if !cfg.Enabled() {
		return c, nil
	}

	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	var ch *amqp091.Channel
	ready := false
	defer func() {
		if ready {
			return
		}
		if ch != nil {
			_ = ch.Close()
		}
		_ = conn.Close()
	}()

	ch, err = conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	// declare main exchange
	if err := ch.ExchangeDeclare(cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("rabbitmq declare exchange: %w", err)
	}

	// declare main queue
	queue, err := ch.QueueDeclare(cfg.Queue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    cfg.DeadLetterExchange,
		"x-dead-letter-routing-key": cfg.DeadLetterRoutingKey,
	})
	if err != nil {
		return nil, fmt.Errorf("rabbitmq declare queue: %w", err)
	}

	// bind main queue to main exchange
	for _, routingKey := range []vo.EventType{vo.EventAvatarUploaded, vo.EventAvatarDeleted} {
		if err := ch.QueueBind(queue.Name, string(routingKey), cfg.Exchange, false, nil); err != nil {
			return nil, fmt.Errorf("rabbitmq bind queue: %w", err)
		}
	}

	// declare dead letter exchange
	if err := ch.ExchangeDeclare(cfg.DeadLetterExchange, deadLetterExchangeType, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("rabbitmq declare dead letter exchange: %w", err)
	}

	// declare dead letter queue
	deadLetterQueue, err := ch.QueueDeclare(cfg.DeadLetterQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq declare dead letter queue: %w", err)
	}

	// bind dead letter queue to dead letter exchange
	if err := ch.QueueBind(deadLetterQueue.Name, cfg.DeadLetterRoutingKey, cfg.DeadLetterExchange, false, nil); err != nil {
		return nil, fmt.Errorf("rabbitmq bind dead letter queue: %w", err)
	}

	// consume messages from main queue
	deliveries, err := ch.Consume(cfg.Queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq consume: %w", err)
	}

	c.conn = conn
	c.ch = ch
	c.deliveries = deliveries
	ready = true
	return c, nil
}

// ReceiveMessages reads avatar lifecycle events from RabbitMQ.
func (c *Consumer) ReceiveMessages(ctx context.Context) (<-chan dto.BrokerMessage, error) {
	if !c.cfg.Enabled() {
		return nil, errBrokerNotConfigured
	}

	if c.deliveries == nil || c.ch == nil || c.conn == nil || c.conn.IsClosed() {
		return nil, errBrokerNotConnected
	}

	messages := make(chan dto.BrokerMessage)

	go func() {
		defer close(messages)

		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-c.deliveries:
				if !ok {
					return
				}

				var (
					msg dto.BrokerMessage
					err error
				)

				switch vo.EventType(delivery.RoutingKey) {
				case vo.EventAvatarUploaded:
					var model model.AvatarUploadedEvent
					if err = easyjson.Unmarshal(delivery.Body, &model); err != nil {
						err = fmt.Errorf("decode avatar uploaded event: %w", err)
						break
					}
					uploaded := c.conv.AvatarUploadedEventModelToAvatarUploadedEventDto(model)
					msg.Uploaded = &uploaded
				case vo.EventAvatarDeleted:
					var model model.AvatarDeletedEvent
					if err = easyjson.Unmarshal(delivery.Body, &model); err != nil {
						err = fmt.Errorf("decode avatar deleted event: %w", err)
						break
					}
					deleted := c.conv.AvatarDeletedEventModelToAvatarDeletedEventDto(model)
					msg.Deleted = &deleted
				default:
					err = fmt.Errorf("rabbitmq: unknown routing key %q", delivery.RoutingKey)
				}

				if err != nil {
					c.log.Error("rabbitmq: skip invalid broker message",
						"error", err,
						"routing_key", delivery.RoutingKey,
					)
					if nackErr := delivery.Nack(false, false); nackErr != nil {
						c.log.Error("rabbitmq: nack invalid broker message failed",
							"error", nackErr,
							"routing_key", delivery.RoutingKey,
						)
					}
					continue
				}

				msg.Delivery = &messageDelivery{delivery: delivery}

				select {
				case messages <- msg:
				case <-ctx.Done():
					if nackErr := delivery.Nack(false, true); nackErr != nil {
						c.log.Error("rabbitmq: requeue message on shutdown failed",
							"error", nackErr,
							"routing_key", delivery.RoutingKey,
						)
					}
					return
				}
			}
		}
	}()

	return messages, nil
}

// Close closes the RabbitMQ connection.
func (c *Consumer) Close() error {
	if !c.cfg.Enabled() {
		return nil
	}

	var closeErr error
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			closeErr = fmt.Errorf("rabbitmq close channel: %w", err)
		}
		c.ch = nil
	}

	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			if closeErr != nil {
				return fmt.Errorf("%v; rabbitmq close connection: %w", closeErr, err)
			}
			return fmt.Errorf("rabbitmq close connection: %w", err)
		}
		c.conn = nil
	}

	c.deliveries = nil
	return closeErr
}

// messageDelivery implements MessageDelivery interface.
type messageDelivery struct {
	delivery amqp091.Delivery
}

// Ack acknowledges the delivery.
func (d *messageDelivery) Ack(_ context.Context) error {
	return d.delivery.Ack(false)
}

// Nack rejects the delivery.
func (d *messageDelivery) Nack(_ context.Context, requeue bool) error {
	return d.delivery.Nack(false, requeue)
}
