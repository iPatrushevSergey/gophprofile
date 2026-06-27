package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// ConfirmAvatarEvent acknowledges or rejects a broker delivery.
type ConfirmAvatarEvent struct{}

// NewConfirmAvatarEvent returns the confirm avatar event use case.
func NewConfirmAvatarEvent() appport.UseCase[dto.ConfirmAvatarEventInput, struct{}] {
	return &ConfirmAvatarEvent{}
}

// Execute confirms broker message delivery.
func (uc *ConfirmAvatarEvent) Execute(ctx context.Context, in dto.ConfirmAvatarEventInput) (struct{}, error) {
	if in.Delivery == nil {
		return struct{}{}, application.ErrBadInput
	}

	if in.Success {
		return struct{}{}, in.Delivery.Ack(ctx)
	}

	return struct{}{}, in.Delivery.Nack(ctx, in.Requeue)
}
