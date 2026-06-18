package port

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

// OutboxWriter persists outbox records and marks them published.
type OutboxWriter interface {
	EnqueueAvatarUploaded(ctx context.Context, event dto.AvatarUploadedEvent) error
	EnqueueAvatarDeleted(ctx context.Context, event dto.AvatarDeletedEvent) error
	MarkPublished(ctx context.Context, id string) error
}

// OutboxReader loads pending outbox records.
type OutboxReader interface {
	ListPending(ctx context.Context, limit int) ([]dto.OutboxEntry, error)
}

// OutboxRepo combines outbox read and write access.
type OutboxRepo interface {
	OutboxWriter
	OutboxReader
}
