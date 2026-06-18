package usecase

import (
	"context"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// PublishPendingOutboxEvents delivers pending outbox records to the message broker.
type PublishPendingOutboxEvents struct {
	outboxReader   appport.OutboxReader
	outboxWriter   appport.OutboxWriter
	eventPublisher appport.EventPublisher
	batchSize      int
}

// NewPublishPendingOutboxEvents returns the publish pending outbox events use case.
func NewPublishPendingOutboxEvents(
	outboxReader appport.OutboxReader,
	outboxWriter appport.OutboxWriter,
	eventPublisher appport.EventPublisher,
	batchSize int,
) appport.UseCase[struct{}, struct{}] {
	return &PublishPendingOutboxEvents{
		outboxReader:   outboxReader,
		outboxWriter:   outboxWriter,
		eventPublisher: eventPublisher,
		batchSize:      batchSize,
	}
}

// Execute publishes pending outbox records.
func (uc *PublishPendingOutboxEvents) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	entries, err := uc.outboxReader.ListPending(ctx, uc.batchSize)
	if err != nil {
		return struct{}{}, err
	}

	for _, entry := range entries {
		if entry.Uploaded != nil {
			if err := uc.eventPublisher.PublishAvatarUploaded(ctx, *entry.Uploaded); err != nil {
				return struct{}{}, err
			}
		}
		if entry.Deleted != nil {
			if err := uc.eventPublisher.PublishAvatarDeleted(ctx, *entry.Deleted); err != nil {
				return struct{}{}, err
			}
		}

		if err := uc.outboxWriter.MarkPublished(ctx, entry.ID); err != nil {
			return struct{}{}, err
		}
	}

	return struct{}{}, nil
}
