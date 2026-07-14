package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// SubscribeAvatarEvents opens a broker event subscription.
type SubscribeAvatarEvents struct {
	eventConsumer appport.EventConsumer
}

// NewSubscribeAvatarEvents returns the subscribe avatar events use case.
func NewSubscribeAvatarEvents(eventConsumer appport.EventConsumer) appport.UseCase[struct{}, <-chan dto.BrokerMessage] {
	return &SubscribeAvatarEvents{eventConsumer: eventConsumer}
}

// Execute opens a broker event stream.
func (uc *SubscribeAvatarEvents) Execute(ctx context.Context, _ struct{}) (<-chan dto.BrokerMessage, error) {
	return uc.eventConsumer.ReceiveMessages(ctx)
}
