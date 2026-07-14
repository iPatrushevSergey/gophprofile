package converter

import (
	"encoding/json"
	"fmt"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
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

	if len(source.TraceCarrier) > 0 {
		var carrier map[string]string
		if err := json.Unmarshal(source.TraceCarrier, &carrier); err != nil {
			return dto.OutboxEvent{}, fmt.Errorf("decode outbox trace carrier: %w", err)
		}
		event.TraceCarrier = carrier
	}

	return event, nil
}

// goverter:converter
// goverter:output:file generated/outbox.go
// goverter:extend CopyTime
// goverter:extend StringToUUID
type OutboxConverter interface {
	// goverter:ignore EventType
	// goverter:ignore Status
	// goverter:ignore PublishedAt
	// goverter:map Event Payload | AvatarUploadedEventToPayload
	// goverter:map TraceCarrier TraceCarrier | TraceCarrierToRawMessage
	OutboxUploadedCreateToOutboxEventModel(source dto.OutboxUploadedCreate) (model.OutboxEvent, error)
	// goverter:ignore EventType
	// goverter:ignore Status
	// goverter:ignore PublishedAt
	// goverter:map Event Payload | AvatarDeletedEventToPayload
	// goverter:map TraceCarrier TraceCarrier | TraceCarrierToRawMessage
	OutboxDeletedCreateToOutboxEventModel(source dto.OutboxDeletedCreate) (model.OutboxEvent, error)
}

// StringToOutboxEventType parses outbox event type from database text.
func StringToOutboxEventType(raw string) vo.OutboxEventType {
	return vo.OutboxEventType(raw)
}

// AvatarUploadedEventToPayload serializes uploaded event payload for outbox storage.
func AvatarUploadedEventToPayload(source dto.AvatarUploadedEvent) (json.RawMessage, error) {
	return json.Marshal(source)
}

// AvatarDeletedEventToPayload serializes deleted event payload for outbox storage.
func AvatarDeletedEventToPayload(source dto.AvatarDeletedEvent) (json.RawMessage, error) {
	return json.Marshal(source)
}

// TraceCarrierToRawMessage serializes trace propagation fields for outbox storage.
func TraceCarrierToRawMessage(source map[string]string) (json.RawMessage, error) {
	if len(source) == 0 {
		return nil, nil
	}

	return json.Marshal(source)
}
