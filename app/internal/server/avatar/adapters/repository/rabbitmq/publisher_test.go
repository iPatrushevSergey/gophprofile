package rabbitmq

import (
	"context"
	"testing"

	oteltelemetry "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	"github.com/stretchr/testify/require"
)

func TestNewPublisher_disabled(t *testing.T) {
	publisher, err := NewPublisher(Config{}, oteltelemetry.NewTracer())
	require.NoError(t, err)

	ctx := context.Background()
	require.ErrorIs(t, publisher.Ping(ctx), errBrokerNotConfigured)
	require.ErrorIs(t, publisher.PublishAvatarUploaded(ctx, dto.AvatarUploadedEvent{}, nil), errBrokerNotConfigured)
}
