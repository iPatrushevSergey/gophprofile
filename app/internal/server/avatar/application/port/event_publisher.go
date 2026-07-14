package port

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

// EventPublisher publishes avatar lifecycle events to a message broker.
type EventPublisher interface {
	PublishAvatarUploaded(ctx context.Context, event dto.AvatarUploadedEvent, traceCarrier map[string]string) error
	PublishAvatarDeleted(ctx context.Context, event dto.AvatarDeletedEvent, traceCarrier map[string]string) error
	Ping(ctx context.Context) error
}
