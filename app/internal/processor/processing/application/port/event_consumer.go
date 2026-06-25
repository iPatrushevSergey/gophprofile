package port

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
)

// EventConsumer consumes avatar lifecycle events from a message broker.
type EventConsumer interface {
	Run(ctx context.Context, handler EventHandler) error
}

// EventHandler handles broker events.
type EventHandler interface {
	HandleAvatarUploaded(ctx context.Context, event dto.AvatarUploadedEvent) error
	HandleAvatarDeleted(ctx context.Context, event dto.AvatarDeletedEvent) error
}
