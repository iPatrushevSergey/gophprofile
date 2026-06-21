package rabbitmq

import (
	"context"
	"testing"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	"github.com/stretchr/testify/require"
)

func TestNewPublisher_disabled(t *testing.T) {
	publisher, err := NewPublisher(Config{})
	require.NoError(t, err)

	ctx := context.Background()
	require.ErrorIs(t, publisher.Ping(ctx), errBrokerNotConfigured)
	require.ErrorIs(t, publisher.PublishAvatarUploaded(ctx, dto.AvatarUploadedEvent{}), errBrokerNotConfigured)
}
