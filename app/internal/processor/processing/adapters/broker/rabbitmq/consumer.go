package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

var errBrokerNotConfigured = errors.New("broker is not configured")

var _ appport.EventConsumer = (*Consumer)(nil)

// Consumer reads avatar lifecycle events from RabbitMQ.
type Consumer struct {
	cfg  Config
	log  appport.Logger
	conv converter.MessageConverter

	mu     sync.Mutex
	conn   *amqp091.Connection
	ch     *amqp091.Channel
	closed bool
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

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConsumeSession(); err != nil {
		return nil, err
	}

	return c, nil
}

// ReceiveMessages reads avatar lifecycle events from RabbitMQ.
func (c *Consumer) ReceiveMessages(ctx context.Context) (<-chan dto.BrokerMessage, error) {
	if !c.cfg.Enabled() {
		return nil, errBrokerNotConfigured
	}

	messages := make(chan dto.BrokerMessage)

	go func() {
		defer close(messages)

		for {
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}

			if err := c.ensureConsumeSession(); err != nil {
				c.mu.Unlock()
				c.log.Error("rabbitmq consume setup failed", "error", err)
				if !wait(ctx, c.cfg.ReconnectInterval) {
					return
				}
				continue
			}

			deliveries, err := c.ch.Consume(c.cfg.Queue, "", false, false, false, false, nil)
			c.mu.Unlock()
			if err != nil {
				c.log.Error("rabbitmq consume failed", "error", err)
				if !wait(ctx, c.cfg.ReconnectInterval) {
					return
				}
				continue
			}

		deliveriesLoop:
			for {
				select {
				case <-ctx.Done():
					return
				case delivery, ok := <-deliveries:
					if !ok {
						c.log.Warn("rabbitmq connection lost, reconnecting")
						if !wait(ctx, c.cfg.ReconnectInterval) {
							return
						}
						break deliveriesLoop
					}

					var (
						msg dto.BrokerMessage
						err error
					)

					switch vo.EventType(delivery.RoutingKey) {
					case vo.EventAvatarUploaded:
						var eventModel model.AvatarUploadedEvent
						if err = easyjson.Unmarshal(delivery.Body, &eventModel); err != nil {
							err = fmt.Errorf("decode avatar uploaded event: %w", err)
							break
						}
						uploaded := c.conv.AvatarUploadedEventModelToAvatarUploadedEventDto(eventModel)
						msg.Uploaded = &uploaded
					case vo.EventAvatarDeleted:
						var eventModel model.AvatarDeletedEvent
						if err = easyjson.Unmarshal(delivery.Body, &eventModel); err != nil {
							err = fmt.Errorf("decode avatar deleted event: %w", err)
							break
						}
						deleted := c.conv.AvatarDeletedEventModelToAvatarDeletedEventDto(eventModel)
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
							c.log.Error("rabbitmq: nack on shutdown failed",
								"error", nackErr,
								"routing_key", delivery.RoutingKey,
							)
						}
						return
					}
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

	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true

	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	return nil
}

// ensureConsumeSession opens or restores the RabbitMQ consume session. Caller must hold c.mu.
func (c *Consumer) ensureConsumeSession() error {
	if c.conn != nil && !c.conn.IsClosed() {
		return nil
	}

	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	conn, err := amqp091.Dial(c.cfg.URL)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
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
		return fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := ch.ExchangeDeclare(c.cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq declare exchange: %w", err)
	}

	queue, err := ch.QueueDeclare(c.cfg.Queue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    c.cfg.DeadLetterExchange,
		"x-dead-letter-routing-key": c.cfg.DeadLetterRoutingKey,
	})
	if err != nil {
		return fmt.Errorf("rabbitmq declare queue: %w", err)
	}

	for _, routingKey := range []vo.EventType{vo.EventAvatarUploaded, vo.EventAvatarDeleted} {
		if err := ch.QueueBind(queue.Name, string(routingKey), c.cfg.Exchange, false, nil); err != nil {
			return fmt.Errorf("rabbitmq bind queue: %w", err)
		}
	}

	if err := ch.ExchangeDeclare(c.cfg.DeadLetterExchange, deadLetterExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq declare dead letter exchange: %w", err)
	}

	deadLetterQueue, err := ch.QueueDeclare(c.cfg.DeadLetterQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq declare dead letter queue: %w", err)
	}

	if err := ch.QueueBind(deadLetterQueue.Name, c.cfg.DeadLetterRoutingKey, c.cfg.DeadLetterExchange, false, nil); err != nil {
		return fmt.Errorf("rabbitmq bind dead letter queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("rabbitmq qos: %w", err)
	}

	c.conn = conn
	c.ch = ch
	ready = true
	return nil
}

// wait waits for a context to be done or a timer to expire.
func wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
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
