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

func TestListUserAvatars_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	avatar := entity.NewAvatar("avatar-1", "user-1", "avatar.png", "image/png", 5, "user-1/avatar-1/original", vo.UploadStatusCompleted, now)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		reader.EXPECT().ListByUserID(ctx, "user-1").Return([]entity.Avatar{*avatar}, nil)

		uc := NewListUserAvatars(reader)
		out, err := uc.Execute(ctx, dto.ListUserAvatarsInput{UserID: "user-1"})
		require.NoError(t, err)
		assert.Len(t, out.Items, 1)
		assert.Equal(t, "avatar-1", out.Items[0].ID)
	})
}
