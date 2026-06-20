package usecase

import (
	"context"
	"fmt"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// PublishPendingOutboxEvents delivers pending outbox records to the message broker.
type PublishPendingOutboxEvents struct {
	outboxReader   appport.OutboxReader
	outboxWriter   appport.OutboxWriter
	eventPublisher appport.EventPublisher
	clock          appport.Clock
	batchSize      int
}

// NewPublishPendingOutboxEvents returns the publish pending outbox events use case.
func NewPublishPendingOutboxEvents(
	outboxReader appport.OutboxReader,
	outboxWriter appport.OutboxWriter,
	eventPublisher appport.EventPublisher,
	clock appport.Clock,
	batchSize int,
) appport.UseCase[struct{}, struct{}] {
	return &PublishPendingOutboxEvents{
		outboxReader:   outboxReader,
		outboxWriter:   outboxWriter,
		eventPublisher: eventPublisher,
		clock:          clock,
		batchSize:      batchSize,
	}
}

// Execute publishes pending outbox records.
func (uc *PublishPendingOutboxEvents) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	events, err := uc.outboxReader.ListPending(ctx, uc.batchSize)
	if err != nil {
		return struct{}{}, err
	}

	for _, event := range events {
		switch event.EventType {
		case vo.OutboxEventAvatarUploaded:
			if err := uc.eventPublisher.PublishAvatarUploaded(ctx, event.Uploaded); err != nil {
				return struct{}{}, err
			}
		case vo.OutboxEventAvatarDeleted:
			if err := uc.eventPublisher.PublishAvatarDeleted(ctx, event.Deleted); err != nil {
				return struct{}{}, err
			}
		default:
			return struct{}{}, fmt.Errorf("unknown outbox event type: %s", event.EventType)
		}

		if err := uc.outboxWriter.MarkPublished(ctx, event.ID, uc.clock.Now()); err != nil {
			return struct{}{}, err
		}
	}

	return struct{}{}, nil
}
