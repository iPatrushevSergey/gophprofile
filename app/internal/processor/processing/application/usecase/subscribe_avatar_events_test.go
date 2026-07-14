package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port/mocks"
)

func TestSubscribeAvatarEvents_Execute(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	consumer := portmocks.NewMockEventConsumer(ctrl)
	ch := make(<-chan dto.BrokerMessage)

	consumer.EXPECT().ReceiveMessages(ctx).Return(ch, nil)

	uc := NewSubscribeAvatarEvents(consumer)
	got, err := uc.Execute(ctx, struct{}{})
	require.NoError(t, err)
	require.Equal(t, ch, got)
}
