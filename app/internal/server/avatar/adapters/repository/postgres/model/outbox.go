package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OutboxEvent is the DB projection of the avatar_outbox table row.
type OutboxEvent struct {
	ID           uuid.UUID
	EventType    string
	Payload      json.RawMessage
	TraceCarrier json.RawMessage
	Status       string
	CreatedAt    time.Time
	PublishedAt  *time.Time
}
