package port

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

// OutboxWriter provides write access to outbox.
type OutboxWriter interface {
	CreateUploaded(ctx context.Context, event dto.OutboxUploadedCreate) error
	CreateDeleted(ctx context.Context, event dto.OutboxDeletedCreate) error
	MarkPublished(ctx context.Context, id string, publishedAt time.Time) error
	ReleaseStalePublishing(ctx context.Context, publishingBefore time.Time) error
}

// OutboxReader provides read access to outbox.
type OutboxReader interface {
	MarkPublishing(ctx context.Context, limit int, publishingAt time.Time) ([]dto.OutboxEvent, error)
}

// OutboxRepo combines outbox read and write access.
type OutboxRepo interface {
	OutboxWriter
	OutboxReader
}
