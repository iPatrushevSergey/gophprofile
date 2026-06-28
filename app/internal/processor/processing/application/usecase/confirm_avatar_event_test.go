package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
)

type stubDelivery struct {
	ackErr  error
	nackErr error
}

func (s stubDelivery) Ack(context.Context) error        { return s.ackErr }
func (s stubDelivery) Nack(context.Context, bool) error { return s.nackErr }

func TestConfirmAvatarEvent_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("badInput", func(t *testing.T) {
		uc := NewConfirmAvatarEvent()
		_, err := uc.Execute(ctx, dto.ConfirmAvatarEventInput{})
		require.ErrorIs(t, err, application.ErrBadInput)
	})

	t.Run("ack", func(t *testing.T) {
		uc := NewConfirmAvatarEvent()
		_, err := uc.Execute(ctx, dto.ConfirmAvatarEventInput{
			Delivery: stubDelivery{},
			Success:  true,
		})
		require.NoError(t, err)
	})
}
