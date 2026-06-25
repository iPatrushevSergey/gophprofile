package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	amqp091 "github.com/rabbitmq/amqp091-go"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

const exchangeType = "topic"

var errBrokerNotConfigured = errors.New("broker is not configured")

var _ appport.EventPublisher = (*Publisher)(nil)

// Publisher publishes avatar events to RabbitMQ.
type Publisher struct {
	cfg Config

	mu   sync.Mutex
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

// NewPublisher creates a RabbitMQ event publisher.
func NewPublisher(cfg Config) (*Publisher, error) {
	p := &Publisher{cfg: cfg}
	if !cfg.Enabled() {
		return p, nil
	}

	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq declare exchange: %w", err)
	}

	p.conn = conn
	p.ch = ch
	return p, nil
}

// PublishAvatarUploaded publishes avatar uploaded event.
func (p *Publisher) PublishAvatarUploaded(ctx context.Context, event dto.AvatarUploadedEvent) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	body, err := converter.AvatarUploadedEventToMessage(event)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	return p.ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		string(vo.OutboxEventAvatarUploaded),
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
}

// PublishAvatarDeleted publishes avatar deleted event.
func (p *Publisher) PublishAvatarDeleted(ctx context.Context, event dto.AvatarDeletedEvent) error {
	if !p.cfg.Enabled() {
		return errBrokerNotConfigured
	}

	body, err := converter.AvatarDeletedEventToMessage(event)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	return p.ch.PublishWithContext(
		ctx,
		p.cfg.Exchange,
		string(vo.OutboxEventAvatarDeleted),
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
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

	err := p.ch.ExchangeDeclarePassive(
		p.cfg.Exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("rabbitmq exchange: %w", err)
	}
	return nil
}
