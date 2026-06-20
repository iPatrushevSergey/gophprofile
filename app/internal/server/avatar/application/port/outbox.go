package port

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

// OutboxWriter persists outbox records and marks them published.
type OutboxWriter interface {
	CreateUploaded(ctx context.Context, event dto.OutboxUploadedCreate) error
	CreateDeleted(ctx context.Context, event dto.OutboxDeletedCreate) error
	MarkPublished(ctx context.Context, id string, publishedAt time.Time) error
}

// OutboxReader loads pending outbox records.
type OutboxReader interface {
	ListPending(ctx context.Context, limit int) ([]dto.OutboxEvent, error)
}

// OutboxRepo combines outbox read and write access.
type OutboxRepo interface {
	OutboxWriter
	OutboxReader
}
