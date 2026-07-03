package dto

import (
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// OutboxEvent is a pending outbox event read from storage.
type OutboxEvent struct {
	ID           string
	EventType    vo.OutboxEventType
	AvatarID     string
	UserID       string
	S3Key        string
	S3Keys       []string
	TraceCarrier map[string]string
}

// OutboxUploadedCreate is an outbox record to persist for avatar upload.
type OutboxUploadedCreate struct {
	ID           string
	CreatedAt    time.Time
	Event        AvatarUploadedEvent
	TraceCarrier map[string]string
}

// OutboxDeletedCreate is an outbox record to persist for avatar delete.
type OutboxDeletedCreate struct {
	ID           string
	CreatedAt    time.Time
	Event        AvatarDeletedEvent
	TraceCarrier map[string]string
}
