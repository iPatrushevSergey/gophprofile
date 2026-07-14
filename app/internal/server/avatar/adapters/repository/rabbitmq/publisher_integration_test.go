//go:build integration

package rabbitmq_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/testutil"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

func TestPublisher_PublishAvatarUploaded(t *testing.T) {
	publisher := testutil.SetupRabbitMQ(t)
	ctx := context.Background()

	require.NoError(t, publisher.Ping(ctx))
	require.NoError(t, publisher.PublishAvatarUploaded(ctx, dto.AvatarUploadedEvent{
		AvatarID: "avatar-1",
		UserID:   "user-1",
		S3Key:    "user-1/avatar-1/original",
	}))
	require.NoError(t, publisher.PublishAvatarDeleted(ctx, dto.AvatarDeletedEvent{
		AvatarID: "avatar-1",
		S3Keys:   []string{"user-1/avatar-1/original"},
	}))
}
