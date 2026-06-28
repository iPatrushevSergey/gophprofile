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

func TestGetAvatar_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	avatar := entity.NewAvatar(
		"avatar-1",
		"user-1",
		"avatar.png",
		"image/png",
		5,
		"user-1/avatar-1/original",
		vo.UploadStatusCompleted,
		now,
	)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)

		reader.EXPECT().FindByID(ctx, "avatar-1").Return(avatar, nil)
		storage.EXPECT().Get(ctx, "user-1/avatar-1/original").Return([]byte("image"), nil)

		uc := NewGetAvatar(reader, storage)
		out, err := uc.Execute(ctx, dto.GetAvatarInput{
			AvatarID:      "avatar-1",
			ThumbnailSize: vo.ThumbnailOriginal,
		})
		require.NoError(t, err)
		assert.Equal(t, []byte("image"), out.Content)
		assert.Equal(t, "image/png", out.MimeType)
	})

	t.Run("usesThumbnailFormat", func(t *testing.T) {
		thumbAvatar := *avatar
		thumbAvatar.ThumbnailS3Keys = map[vo.ThumbnailSize]map[vo.OutputFormat]string{
			vo.ThumbnailSize100: {
				vo.OutputFormatWebP: "user-1/avatar-1/100x100/webp",
			},
		}

		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)

		reader.EXPECT().FindByID(ctx, "avatar-1").Return(&thumbAvatar, nil)
		storage.EXPECT().Get(ctx, "user-1/avatar-1/100x100/webp").Return([]byte("thumb"), nil)

		uc := NewGetAvatar(reader, storage)
		out, err := uc.Execute(ctx, dto.GetAvatarInput{
			AvatarID:      "avatar-1",
			ThumbnailSize: vo.ThumbnailSize100,
			OutputFormat:  vo.OutputFormatWebP,
		})
		require.NoError(t, err)
		assert.Equal(t, []byte("thumb"), out.Content)
		assert.Equal(t, "image/webp", out.MimeType)
	})

	t.Run("thumbnailNotReady", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		reader.EXPECT().FindByID(ctx, "avatar-1").Return(avatar, nil)

		uc := NewGetAvatar(reader, portmocks.NewMockAvatarStorage(ctrl))
		_, err := uc.Execute(ctx, dto.GetAvatarInput{
			AvatarID:      "avatar-1",
			ThumbnailSize: vo.ThumbnailSize100,
		})
		assert.ErrorIs(t, err, application.ErrNotFound)
	})

	t.Run("notFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		reader.EXPECT().FindByID(ctx, "missing").Return(nil, application.ErrNotFound)

		uc := NewGetAvatar(reader, portmocks.NewMockAvatarStorage(ctrl))
		_, err := uc.Execute(ctx, dto.GetAvatarInput{
			AvatarID:      "missing",
			ThumbnailSize: vo.ThumbnailOriginal,
		})
		assert.ErrorIs(t, err, application.ErrNotFound)
	})
}
