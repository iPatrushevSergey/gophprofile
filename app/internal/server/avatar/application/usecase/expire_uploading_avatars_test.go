package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestExpireUploadingAvatars_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	avatar := entity.NewAvatar("avatar-1", "user-1", "avatar.png", "image/png", 5, "user-1/avatar-1/original", vo.UploadStatusUploading, now)
	other := entity.NewAvatar("avatar-2", "user-1", "avatar2.png", "image/png", 5, "user-1/avatar-2/original", vo.UploadStatusUploading, now)

	t.Run("expiresUploading", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		clock.EXPECT().Now().Return(now)
		reader.EXPECT().ListExpiredUploading(ctx, now.Add(-30*time.Minute)).Return([]entity.Avatar{*avatar}, nil)
		storage.EXPECT().Delete(ctx, "user-1/avatar-1/original").Return(nil)
		writer.EXPECT().MarkUploadFailed(ctx, "avatar-1", now).Return(nil)

		uc := NewExpireUploadingAvatars(reader, writer, storage, clock, logger.NewNopLogger(), 30*time.Minute)
		_, err := uc.Execute(ctx, struct{}{})
		require.NoError(t, err)
	})

	t.Run("continuesWhenDeleteFails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		clock.EXPECT().Now().Return(now)
		reader.EXPECT().ListExpiredUploading(ctx, now.Add(-30*time.Minute)).Return([]entity.Avatar{*avatar, *other}, nil)
		storage.EXPECT().Delete(ctx, "user-1/avatar-1/original").Return(errors.New("storage unavailable"))
		storage.EXPECT().Delete(ctx, "user-1/avatar-2/original").Return(nil)
		writer.EXPECT().MarkUploadFailed(ctx, "avatar-2", now).Return(nil)

		uc := NewExpireUploadingAvatars(reader, writer, storage, clock, logger.NewNopLogger(), 30*time.Minute)
		_, err := uc.Execute(ctx, struct{}{})
		require.NoError(t, err)
	})
}
