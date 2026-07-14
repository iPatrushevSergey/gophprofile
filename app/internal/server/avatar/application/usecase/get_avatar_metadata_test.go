package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestGetAvatarMetadata_Execute(t *testing.T) {
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
	avatar.Width = 1920
	avatar.Height = 1080
	avatar.ProcessingStatus = vo.ProcessingStatusCompleted

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		reader.EXPECT().FindByID(ctx, "avatar-1").Return(avatar, nil)

		uc := NewGetAvatarMetadata(reader)
		out, err := uc.Execute(ctx, dto.GetAvatarMetadataInput{AvatarID: "avatar-1"})
		require.NoError(t, err)
		assert.Equal(t, "avatar-1", out.ID)
		assert.Equal(t, 1920, out.Width)
		assert.Equal(t, 1080, out.Height)
		assert.Equal(t, vo.ProcessingStatusCompleted, out.ProcessingStatus)
	})
}
