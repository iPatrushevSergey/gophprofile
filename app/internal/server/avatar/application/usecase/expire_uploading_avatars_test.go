package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestExpireUploadingAvatars_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	avatar := entity.NewAvatar("avatar-1", "user-1", "avatar.png", "image/png", 5, "user-1/avatar-1/original", vo.UploadStatusUploading, now)

	t.Run("expiresUploading", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		clock.EXPECT().Now().Return(now)
		reader.EXPECT().ListExpiredUploading(ctx, now.Add(-30*time.Minute)).Return([]entity.Avatar{*avatar}, nil)
		storage.EXPECT().Delete(ctx, "user-1/avatar-1/original").Return(nil)
		writer.EXPECT().MarkUploadFailed(ctx, "avatar-1").Return(nil)

		uc := NewExpireUploadingAvatars(reader, writer, storage, clock, 30*time.Minute)
		_, err := uc.Execute(ctx, struct{}{})
		require.NoError(t, err)
	})

	t.Run("returnsErrorWhenDeleteFails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		clock.EXPECT().Now().Return(now)
		reader.EXPECT().ListExpiredUploading(ctx, now.Add(-30*time.Minute)).Return([]entity.Avatar{*avatar}, nil)
		storage.EXPECT().Delete(ctx, "user-1/avatar-1/original").Return(errors.New("storage unavailable"))

		uc := NewExpireUploadingAvatars(reader, writer, storage, clock, 30*time.Minute)
		_, err := uc.Execute(ctx, struct{}{})
		require.EqualError(t, err, "storage unavailable")
	})
}
