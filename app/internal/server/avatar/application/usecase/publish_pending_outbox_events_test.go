package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestPublishPendingOutboxEvents_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockOutboxReader(ctrl)
		writer := portmocks.NewMockOutboxWriter(ctrl)
		publisher := portmocks.NewMockEventPublisher(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		uploaded := dto.AvatarUploadedEvent{
			AvatarID: "avatar-1",
			UserID:   "user-1",
			S3Key:    "user-1/avatar-1/original",
		}

		reader.EXPECT().ListPending(ctx, 10).Return([]dto.OutboxEvent{
			{
				ID:        "outbox-1",
				EventType: vo.OutboxEventAvatarUploaded,
				Uploaded:  uploaded,
			},
		}, nil)
		publisher.EXPECT().PublishAvatarUploaded(ctx, uploaded).Return(nil)
		clock.EXPECT().Now().Return(now)
		writer.EXPECT().MarkPublished(ctx, "outbox-1", now).Return(nil)

		uc := NewPublishPendingOutboxEvents(reader, writer, publisher, clock, 10)
		_, err := uc.Execute(ctx, struct{}{})
		require.NoError(t, err)
	})
}
