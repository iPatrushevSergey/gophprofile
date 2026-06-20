package dto

import (
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// OutboxEntry is a pending outbox record with a typed event payload.
type OutboxEntry struct {
	ID        string
	EventType vo.OutboxEventType
	Uploaded  *AvatarUploadedEvent
	Deleted   *AvatarDeletedEvent
}
