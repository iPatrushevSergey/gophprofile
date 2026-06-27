package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	easyjson "github.com/mailru/easyjson"
	amqp091 "github.com/rabbitmq/amqp091-go"

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
	cfg  Config
	conv converter.MessageConverter

	mu              sync.Mutex
	conn            *amqp091.Connection
	ch              *amqp091.Channel
	publishConfirms <-chan amqp091.Confirmation
	publishReturns  <-chan amqp091.Return
}

// NewPublisher creates a RabbitMQ event publisher.
func NewPublisher(cfg Config) (*Publisher, error) {
	p := &Publisher{
		cfg:  cfg,
		conv: &generated.MessageConverterImpl{},
	}
	if !cfg.Enabled() {
		return p, nil
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

	if err := ch.ExchangeDeclare(cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("rabbitmq declare exchange: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("rabbitmq enable publisher confirms: %w", err)
	}

	p.conn = conn
	p.ch = ch
	p.publishConfirms = ch.NotifyPublish(make(chan amqp091.Confirmation, 1))
	p.publishReturns = ch.NotifyReturn(make(chan amqp091.Return, 1))
	ready = true
	return p, nil
}

// PublishAvatarUploaded publishes avatar uploaded event.
func (p *Publisher) PublishAvatarUploaded(ctx context.Context, event dto.AvatarUploadedEvent) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	model := p.conv.AvatarUploadedEventDtoToAvatarUploadedEventModel(event)
	body, err := easyjson.Marshal(model)
	if err != nil {
		return fmt.Errorf("encode avatar uploaded event: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		string(vo.OutboxEventAvatarUploaded),
		true,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("rabbitmq publish: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ret := <-p.publishReturns:
			return fmt.Errorf(
				"rabbitmq unroutable message: exchange=%s routing_key=%s reply_code=%d reply_text=%s",
				ret.Exchange, ret.RoutingKey, ret.ReplyCode, ret.ReplyText,
			)
		case confirm := <-p.publishConfirms:
			if !confirm.Ack {
				return fmt.Errorf("rabbitmq publish not acknowledged")
			}
			return nil
		}
	}
}

// PublishAvatarDeleted publishes avatar deleted event.
func (p *Publisher) PublishAvatarDeleted(ctx context.Context, event dto.AvatarDeletedEvent) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	model := p.conv.AvatarDeletedEventDtoToAvatarDeletedEventModel(event)
	body, err := easyjson.Marshal(model)
	if err != nil {
		return fmt.Errorf("encode avatar deleted event: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		string(vo.OutboxEventAvatarDeleted),
		true,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("rabbitmq publish: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ret := <-p.publishReturns:
			return fmt.Errorf(
				"rabbitmq unroutable message: exchange=%s routing_key=%s reply_code=%d reply_text=%s",
				ret.Exchange, ret.RoutingKey, ret.ReplyCode, ret.ReplyText,
			)
		case confirm := <-p.publishConfirms:
			if !confirm.Ack {
				return fmt.Errorf("rabbitmq publish not acknowledged")
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

	if p.ch == nil || p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("rabbitmq: not connected")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := p.ch.ExchangeDeclarePassive(p.cfg.Exchange, exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq exchange: %w", err)
	}

	return nil
}
