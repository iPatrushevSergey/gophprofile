package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	easyjson "github.com/mailru/easyjson"
	amqp091 "github.com/rabbitmq/amqp091-go"

	oteltelemetry "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq/converter/generated"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

const exchangeType = "topic"

var errBrokerNotConfigured = errors.New("broker is not configured")

var _ appport.EventPublisher = (*Publisher)(nil)

// Publisher publishes avatar events to RabbitMQ.
type Publisher struct {
	cfg    Config
	conv   converter.MessageConverter
	tracer pkgport.Tracer

	mu              sync.Mutex
	conn            *amqp091.Connection
	ch              *amqp091.Channel
	publishConfirms <-chan amqp091.Confirmation
	publishReturns  <-chan amqp091.Return
}

// NewPublisher creates a RabbitMQ event publisher.
func NewPublisher(cfg Config, tracer pkgport.Tracer) (*Publisher, error) {
	p := &Publisher{
		cfg:    cfg,
		conv:   &generated.MessageConverterImpl{},
		tracer: tracer,
	}
	if !cfg.Enabled() {
		return p, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ensurePublishSession(); err != nil {
		return nil, err
	}

	return p, nil
}

// PublishAvatarUploaded publishes avatar uploaded event.
func (p *Publisher) PublishAvatarUploaded(ctx context.Context, event dto.AvatarUploadedEvent, traceCarrier map[string]string) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	ctx = p.tracer.MapToContext(ctx, traceCarrier)

	model := p.conv.AvatarUploadedEventDtoToAvatarUploadedEventModel(event)
	body, err := easyjson.Marshal(model)
	if err != nil {
		return fmt.Errorf("encode avatar uploaded event: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ensurePublishSession(); err != nil {
		return err
	}

	// Start a span for the publish operation.
	ctx, span := p.tracer.Start(ctx, pkgport.SpanConfig{
		Key:  "avatar.adapter.rabbitmq_publisher.publish_avatar_uploaded",
		Name: string(vo.OutboxEventAvatarUploaded) + " send",
		Kind: pkgport.SpanKindProducer,
		Attributes: []pkgport.Attribute{
			{Key: "messaging.system", Value: "rabbitmq"},
			{Key: "messaging.destination.name", Value: p.cfg.Exchange},
			{Key: "messaging.rabbitmq.destination.routing_key", Value: string(vo.OutboxEventAvatarUploaded)},
			{Key: "messaging.operation.type", Value: "send"},
		},
	})
	headers := oteltelemetry.InjectAMQP(ctx, amqp091.Table{})
	var publishErr error
	defer func() {
		span.Fail(publishErr)
		span.End()
	}()

	// Publish the message to RabbitMQ.
	if publishErr = p.ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		string(vo.OutboxEventAvatarUploaded),
		true,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Headers:      headers,
			Body:         body,
		},
	); publishErr != nil {
		return fmt.Errorf("rabbitmq publish: %w", publishErr)
	}

	// Wait for the publish confirmation.
	for {
		select {
		case <-ctx.Done():
			publishErr = ctx.Err()
			return publishErr
		case ret := <-p.publishReturns:
			publishErr = fmt.Errorf(
				"rabbitmq unroutable message: exchange=%s routing_key=%s reply_code=%d reply_text=%s",
				ret.Exchange, ret.RoutingKey, ret.ReplyCode, ret.ReplyText,
			)
			return publishErr
		case confirm := <-p.publishConfirms:
			if !confirm.Ack {
				publishErr = fmt.Errorf("rabbitmq publish not acknowledged")
				return publishErr
			}
			return nil
		}
	}
}

// PublishAvatarDeleted publishes avatar deleted event.
func (p *Publisher) PublishAvatarDeleted(ctx context.Context, event dto.AvatarDeletedEvent, traceCarrier map[string]string) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	ctx = p.tracer.MapToContext(ctx, traceCarrier)

	model := p.conv.AvatarDeletedEventDtoToAvatarDeletedEventModel(event)
	body, err := easyjson.Marshal(model)
	if err != nil {
		return fmt.Errorf("encode avatar deleted event: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ensurePublishSession(); err != nil {
		return err
	}

	// Start a span for the publish operation.
	ctx, span := p.tracer.Start(ctx, pkgport.SpanConfig{
		Key:  "avatar.adapter.rabbitmq_publisher.publish_avatar_deleted",
		Name: string(vo.OutboxEventAvatarDeleted) + " send",
		Kind: pkgport.SpanKindProducer,
		Attributes: []pkgport.Attribute{
			{Key: "messaging.system", Value: "rabbitmq"},
			{Key: "messaging.destination.name", Value: p.cfg.Exchange},
			{Key: "messaging.rabbitmq.destination.routing_key", Value: string(vo.OutboxEventAvatarDeleted)},
			{Key: "messaging.operation.type", Value: "send"},
		},
	})
	headers := oteltelemetry.InjectAMQP(ctx, amqp091.Table{})
	var publishErr error
	defer func() {
		span.Fail(publishErr)
		span.End()
	}()

	// Publish the message to RabbitMQ.
	if publishErr = p.ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		string(vo.OutboxEventAvatarDeleted),
		true,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Headers:      headers,
			Body:         body,
		},
	); publishErr != nil {
		return fmt.Errorf("rabbitmq publish: %w", publishErr)
	}

	// Wait for the publish confirmation.
	for {
		select {
		case <-ctx.Done():
			publishErr = ctx.Err()
			return publishErr
		case ret := <-p.publishReturns:
			publishErr = fmt.Errorf(
				"rabbitmq unroutable message: exchange=%s routing_key=%s reply_code=%d reply_text=%s",
				ret.Exchange, ret.RoutingKey, ret.ReplyCode, ret.ReplyText,
			)
			return publishErr
		case confirm := <-p.publishConfirms:
			if !confirm.Ack {
				publishErr = fmt.Errorf("rabbitmq publish not acknowledged")
				return publishErr
			}
			return nil
		}
	}
}

// Ping checks broker connectivity.
func (p *Publisher) Ping(ctx context.Context) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := p.ensurePublishSession(); err != nil {
		return err
	}

	if err := p.ch.ExchangeDeclarePassive(p.cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq exchange: %w", err)
	}

	return nil
}

// ensurePublishSession opens or restores the RabbitMQ publish session.
func (p *Publisher) ensurePublishSession() error {
	if p.conn != nil && !p.conn.IsClosed() && p.ch != nil && !p.ch.IsClosed() {
		return nil
	}

	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}

	conn, err := amqp091.Dial(p.cfg.URL)
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

	if err := ch.ExchangeDeclarePassive(p.cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq exchange %q: %w", p.cfg.Exchange, err)
	}

	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("rabbitmq enable publisher confirms: %w", err)
	}

	p.conn = conn
	p.ch = ch
	p.publishConfirms = ch.NotifyPublish(make(chan amqp091.Confirmation, 1))
	p.publishReturns = ch.NotifyReturn(make(chan amqp091.Return, 1))
	ready = true
	return nil
}
