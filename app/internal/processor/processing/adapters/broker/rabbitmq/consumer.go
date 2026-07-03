package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp091 "github.com/rabbitmq/amqp091-go"

	oteltelemetry "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/converter/generated"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
	easyjson "github.com/mailru/easyjson"
)

const (
	exchangeType             = "topic"
	deadLetterExchangeType   = "direct"
	retryExchangeType        = "direct"
	retryCountHeader         = "x-retry-count"
	originalRoutingKeyHeader = "x-original-routing-key"
)

var errBrokerNotConfigured = errors.New("broker is not configured")

var _ appport.EventConsumer = (*Consumer)(nil)

// Consumer reads avatar lifecycle events from RabbitMQ.
type Consumer struct {
	cfg    Config
	log    pkgport.Logger
	conv   converter.MessageConverter
	tracer pkgport.Tracer

	mu              sync.Mutex
	conn            *amqp091.Connection
	ch              *amqp091.Channel
	publishConfirms <-chan amqp091.Confirmation
	closed          bool
}

// NewConsumer creates a RabbitMQ event consumer.
func NewConsumer(cfg Config, log pkgport.Logger, tracer pkgport.Tracer) (*Consumer, error) {
	c := &Consumer{
		cfg:    cfg,
		log:    log,
		conv:   &generated.MessageConverterImpl{},
		tracer: tracer,
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
			if err != nil {
				c.mu.Unlock()
				c.log.Error("rabbitmq consume failed", "error", err)
				if !wait(ctx, c.cfg.ReconnectInterval) {
					return
				}
				continue
			}
			c.mu.Unlock()

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

					attempt := 1
					if count, ok := delivery.Headers[retryCountHeader].(int32); ok {
						attempt = int(count) + 1
					}

					msgCtx := oteltelemetry.ExtractAMQP(ctx, delivery.Headers)
					routingKey := delivery.RoutingKey
					if routingKey == c.cfg.RetryReturnRoutingKey {
						if original, ok := delivery.Headers[originalRoutingKeyHeader].(string); ok && original != "" {
							routingKey = original
						}
					}
					msgCtx, receiveSpan := c.tracer.Start(msgCtx, pkgport.SpanConfig{
						Key:  "processing.adapter.consumer.receive_messages",
						Name: routingKey + " receive",
						Kind: pkgport.SpanKindConsumer,
						Attributes: []pkgport.Attribute{
							{Key: "messaging.system", Value: "rabbitmq"},
							{Key: "messaging.destination.name", Value: c.cfg.Queue},
							{Key: "messaging.rabbitmq.destination.routing_key", Value: routingKey},
							{Key: "messaging.operation.type", Value: "receive"},
						},
					})

					msg, _, err := c.decodeBrokerMessage(delivery)
					if err != nil {
						receiveSpan.Fail(err)
						receiveSpan.End()

						c.log.Error("rabbitmq: skip invalid broker message",
							"error", err,
							"routing_key", routingKey,
						)

						c.mu.Lock()
						nackErr := delivery.Nack(false, false)
						c.mu.Unlock()

						if nackErr != nil {
							c.log.Error("rabbitmq: nack invalid broker message failed",
								"error", nackErr,
								"routing_key", routingKey,
							)
						}
						continue
					}

					msg.Ctx = msgCtx

					msg.Delivery = &messageDelivery{
						consumer: c,
						delivery: delivery,
						attempt:  attempt,
					}

					select {
					case messages <- msg:
						receiveSpan.End()
					case <-ctx.Done():
						receiveSpan.Fail(ctx.Err())
						receiveSpan.End()
						c.mu.Lock()
						nackErr := delivery.Nack(false, true)
						c.mu.Unlock()

						if nackErr != nil {
							c.log.Error("rabbitmq: nack on shutdown failed",
								"error", nackErr,
								"routing_key", routingKey,
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

// decodeBrokerMessage decodes a broker message from a RabbitMQ delivery.
func (c *Consumer) decodeBrokerMessage(delivery amqp091.Delivery) (dto.BrokerMessage, string, error) {
	routingKey := delivery.RoutingKey
	if routingKey == c.cfg.RetryReturnRoutingKey {
		if original, ok := delivery.Headers[originalRoutingKeyHeader].(string); ok && original != "" {
			routingKey = original
		}
	}

	var msg dto.BrokerMessage

	switch vo.EventType(routingKey) {
	case vo.EventAvatarUploaded:
		var eventModel model.AvatarUploadedEvent
		if err := easyjson.Unmarshal(delivery.Body, &eventModel); err != nil {
			return dto.BrokerMessage{}, routingKey, fmt.Errorf("decode avatar uploaded event: %w", err)
		}
		uploaded := c.conv.AvatarUploadedEventModelToAvatarUploadedEventDto(eventModel)
		msg.Uploaded = &uploaded
	case vo.EventAvatarDeleted:
		var eventModel model.AvatarDeletedEvent
		if err := easyjson.Unmarshal(delivery.Body, &eventModel); err != nil {
			return dto.BrokerMessage{}, routingKey, fmt.Errorf("decode avatar deleted event: %w", err)
		}
		deleted := c.conv.AvatarDeletedEventModelToAvatarDeletedEventDto(eventModel)
		msg.Deleted = &deleted
	default:
		return dto.BrokerMessage{}, routingKey, fmt.Errorf("rabbitmq: unknown routing key %q", routingKey)
	}

	return msg, routingKey, nil
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
		c.publishConfirms = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	return nil
}

// ensureConsumeSession opens or restores the RabbitMQ consume session. Caller must hold c.mu.
func (c *Consumer) ensureConsumeSession() error {
	if c.conn != nil && !c.conn.IsClosed() && c.ch != nil && !c.ch.IsClosed() {
		return nil
	}

	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
		c.publishConfirms = nil
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

	if err := ch.ExchangeDeclarePassive(c.cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq exchange %q: %w", c.cfg.Exchange, err)
	}

	if err := ch.ExchangeDeclarePassive(c.cfg.DeadLetterExchange, deadLetterExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq dead letter exchange %q: %w", c.cfg.DeadLetterExchange, err)
	}

	if err := ch.ExchangeDeclarePassive(c.cfg.RetryExchange, retryExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq retry exchange %q: %w", c.cfg.RetryExchange, err)
	}

	if _, err := ch.QueueDeclarePassive(c.cfg.Queue, true, false, false, false, amqp091.Table{
		"x-dead-letter-exchange":    c.cfg.DeadLetterExchange,
		"x-dead-letter-routing-key": c.cfg.DeadLetterRoutingKey,
	}); err != nil {
		return fmt.Errorf("rabbitmq queue %q: %w", c.cfg.Queue, err)
	}

	if _, err := ch.QueueDeclarePassive(c.cfg.DeadLetterQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq dead letter queue %q: %w", c.cfg.DeadLetterQueue, err)
	}

	retryTTLMillis := int32(c.cfg.RetryTTL / time.Millisecond)

	if _, err := ch.QueueDeclarePassive(c.cfg.RetryQueue, true, false, false, false, amqp091.Table{
		"x-message-ttl":             retryTTLMillis,
		"x-dead-letter-exchange":    c.cfg.RetryExchange,
		"x-dead-letter-routing-key": c.cfg.RetryReturnRoutingKey,
	}); err != nil {
		return fmt.Errorf("rabbitmq retry queue %q: %w", c.cfg.RetryQueue, err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("rabbitmq qos: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("rabbitmq enable publisher confirms: %w", err)
	}

	c.conn = conn
	c.ch = ch
	c.publishConfirms = ch.NotifyPublish(make(chan amqp091.Confirmation, 1))
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
	consumer *Consumer
	delivery amqp091.Delivery
	attempt  int
}

// Ack acknowledges the delivery.
func (d *messageDelivery) Ack(_ context.Context) error {
	d.consumer.mu.Lock()
	defer d.consumer.mu.Unlock()

	return d.delivery.Ack(false)
}

// Nack rejects the delivery.
func (d *messageDelivery) Nack(ctx context.Context, requeue bool) error {
	d.consumer.mu.Lock()
	defer d.consumer.mu.Unlock()

	if !requeue {
		return d.delivery.Nack(false, false)
	}

	if d.attempt >= d.consumer.cfg.MaxRetries {
		d.consumer.log.Warn("rabbitmq: max delivery attempts reached, sending to dlq",
			"routing_key", d.delivery.RoutingKey,
			"attempt", d.attempt,
			"max_retries", d.consumer.cfg.MaxRetries,
		)
		return d.delivery.Nack(false, false)
	}

	headers := amqp091.Table{
		retryCountHeader: int32(d.attempt),
	}
	routingKey := d.delivery.RoutingKey
	if routingKey == d.consumer.cfg.RetryReturnRoutingKey {
		if original, ok := d.delivery.Headers[originalRoutingKeyHeader].(string); ok && original != "" {
			routingKey = original
		}
	}
	if routingKey != "" {
		headers[originalRoutingKeyHeader] = routingKey
	}

	if err := d.consumer.ensureConsumeSession(); err != nil {
		return fmt.Errorf("rabbitmq republish for retry: %w", err)
	}

	if err := d.consumer.ch.PublishWithContext(
		ctx,
		"",
		d.consumer.cfg.RetryQueue,
		false,
		false,
		amqp091.Publishing{
			ContentType:  d.delivery.ContentType,
			DeliveryMode: amqp091.Persistent,
			Body:         d.delivery.Body,
			Headers:      headers,
		},
	); err != nil {
		return fmt.Errorf("rabbitmq republish for retry: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case confirm := <-d.consumer.publishConfirms:
			if !confirm.Ack {
				return fmt.Errorf("rabbitmq republish for retry not acknowledged")
			}
			if err := d.delivery.Ack(false); err != nil {
				return fmt.Errorf("rabbitmq ack after retry republish: %w", err)
			}
			return nil
		}
	}
}
