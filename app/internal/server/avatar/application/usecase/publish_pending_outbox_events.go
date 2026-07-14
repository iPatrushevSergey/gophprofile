package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// PublishPendingOutboxEvents delivers pending outbox records to the message broker.
type PublishPendingOutboxEvents struct {
	outboxReader      appport.OutboxReader
	outboxWriter      appport.OutboxWriter
	eventPublisher    appport.EventPublisher
	clock             appport.Clock
	batchSize         int
	publishingTimeout time.Duration
}

// NewPublishPendingOutboxEvents returns the publish pending outbox events use case.
func NewPublishPendingOutboxEvents(
	outboxReader appport.OutboxReader,
	outboxWriter appport.OutboxWriter,
	eventPublisher appport.EventPublisher,
	clock appport.Clock,
	batchSize int,
	publishingTimeout time.Duration,
) appport.UseCase[struct{}, struct{}] {
	return &PublishPendingOutboxEvents{
		outboxReader:      outboxReader,
		outboxWriter:      outboxWriter,
		eventPublisher:    eventPublisher,
		clock:             clock,
		batchSize:         batchSize,
		publishingTimeout: publishingTimeout,
	}
}

// Execute publishes pending outbox records.
func (uc *PublishPendingOutboxEvents) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	if err := uc.outboxWriter.ReleaseStalePublishing(ctx, uc.clock.Now().Add(-uc.publishingTimeout)); err != nil {
		return struct{}{}, err
	}

	events, err := uc.outboxReader.MarkPublishing(ctx, uc.batchSize, uc.clock.Now())
	if err != nil {
		return struct{}{}, err
	}

	for _, event := range events {
		switch event.EventType {
		case vo.OutboxEventAvatarUploaded:
			if err := uc.eventPublisher.PublishAvatarUploaded(ctx, dto.AvatarUploadedEvent{
				AvatarID: event.AvatarID,
				UserID:   event.UserID,
				S3Key:    event.S3Key,
			}, event.TraceCarrier); err != nil {
				return struct{}{}, err
			}
		case vo.OutboxEventAvatarDeleted:
			if err := uc.eventPublisher.PublishAvatarDeleted(ctx, dto.AvatarDeletedEvent{
				AvatarID: event.AvatarID,
				S3Keys:   event.S3Keys,
			}, event.TraceCarrier); err != nil {
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
