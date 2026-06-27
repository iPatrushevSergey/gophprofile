package port

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
)

// EventConsumer reads avatar lifecycle events from a message broker.
type EventConsumer interface {
	ReceiveMessages(ctx context.Context) (<-chan dto.BrokerMessage, error)
	Close() error
}
