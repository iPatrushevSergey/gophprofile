package converter

import (
	"encoding/json"
	"fmt"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// OutboxEventModelToOutboxEventDto maps a stored outbox row to an application event.
func OutboxEventModelToOutboxEventDto(source model.OutboxEvent) (dto.OutboxEvent, error) {
	event := dto.OutboxEvent{
		ID:        UUIDToString(source.ID),
		EventType: StringToOutboxEventType(source.EventType),
	}

	switch event.EventType {
	case vo.OutboxEventAvatarUploaded:
		var uploaded dto.AvatarUploadedEvent
		if err := json.Unmarshal(source.Payload, &uploaded); err != nil {
			return dto.OutboxEvent{}, fmt.Errorf("decode avatar uploaded outbox payload: %w", err)
		}
		event.AvatarID = uploaded.AvatarID
		event.UserID = uploaded.UserID
		event.S3Key = uploaded.S3Key
	case vo.OutboxEventAvatarDeleted:
		var deleted dto.AvatarDeletedEvent
		if err := json.Unmarshal(source.Payload, &deleted); err != nil {
			return dto.OutboxEvent{}, fmt.Errorf("decode avatar deleted outbox payload: %w", err)
		}
		event.AvatarID = deleted.AvatarID
		event.S3Keys = deleted.S3Keys
	default:
		return dto.OutboxEvent{}, fmt.Errorf("unknown outbox event type: %s", event.EventType)
	}

	return event, nil
}

// goverter:converter
// goverter:output:file generated/outbox.go
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:extend StringToUUID
// goverter:extend OutboxEventTypeToString
// goverter:extend OutboxStatusToString
type OutboxConverter interface {
	// goverter:map . Payload | OutboxEntityToPayload
	OutboxEntityToOutboxEventModel(source entity.OutboxEvent) (model.OutboxEvent, error)
}

// OutboxEventTypeToString converts outbox event type to database text.
func OutboxEventTypeToString(eventType vo.OutboxEventType) string {
	return string(eventType)
}

// StringToOutboxEventType parses outbox event type from database text.
func StringToOutboxEventType(raw string) vo.OutboxEventType {
	return vo.OutboxEventType(raw)
}

// OutboxStatusToString converts outbox status to database text.
func OutboxStatusToString(status vo.OutboxStatus) string {
	return string(status)
}

// OutboxEntityToPayload serializes outbox entity payload for storage.
func OutboxEntityToPayload(source entity.OutboxEvent) (json.RawMessage, error) {
	switch source.EventType {
	case vo.OutboxEventAvatarUploaded:
		return json.Marshal(dto.AvatarUploadedEvent{
			AvatarID: source.AvatarID,
			UserID:   source.UserID,
			S3Key:    source.S3Key,
		})
	case vo.OutboxEventAvatarDeleted:
		return json.Marshal(dto.AvatarDeletedEvent{
			AvatarID: source.AvatarID,
			S3Keys:   source.S3Keys,
		})
	default:
		return nil, fmt.Errorf("unknown outbox event type: %s", source.EventType)
	}
}
