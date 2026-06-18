package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestUploadAvatar_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		outbox := portmocks.NewMockOutboxWriter(ctrl)
		transactor := portmocks.NewMockTransactor(ctrl)
		idGen := portmocks.NewMockIDGenerator(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		idGen.EXPECT().NewID().Return("avatar-1", nil)
		clock.EXPECT().Now().Return(now)
		storage.EXPECT().Put(ctx, "user-1/avatar-1/original", []byte("image"), "image/png").Return(nil)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
		)
		writer.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, avatar *entity.Avatar) error {
			assert.Equal(t, "avatar-1", avatar.ID)
			assert.Equal(t, "user-1", avatar.UserID)
			assert.Equal(t, vo.UploadStatusCompleted, avatar.UploadStatus)
			assert.Equal(t, vo.ProcessingStatusPending, avatar.ProcessingStatus)
			return nil
		})
		outbox.EXPECT().EnqueueAvatarUploaded(ctx, dto.AvatarUploadedEvent{
			AvatarID: "avatar-1",
			UserID:   "user-1",
			S3Key:    "user-1/avatar-1/original",
		}).Return(nil)

		uc := NewUploadAvatar(writer, storage, outbox, transactor, idGen, clock)
		out, err := uc.Execute(ctx, dto.UploadAvatarInput{
			UserID:   "user-1",
			FileName: "avatar.png",
			MimeType: "image/png",
			Content:  []byte("image"),
		})
		require.NoError(t, err)
		assert.Equal(t, "avatar-1", out.ID)
		assert.Equal(t, vo.ProcessingStatusPending, out.ProcessingStatus)
	})

	t.Run("invalidMimeType", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		uc := NewUploadAvatar(
			portmocks.NewMockAvatarWriter(ctrl),
			portmocks.NewMockAvatarStorage(ctrl),
			portmocks.NewMockOutboxWriter(ctrl),
			portmocks.NewMockTransactor(ctrl),
			portmocks.NewMockIDGenerator(ctrl),
			portmocks.NewMockClock(ctrl),
		)

		_, err := uc.Execute(ctx, dto.UploadAvatarInput{
			UserID:   "user-1",
			FileName: "avatar.png",
			MimeType: "text/plain",
			Content:  []byte("image"),
		})
		assert.ErrorIs(t, err, application.ErrBadInput)
	})
}
