package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port/mocks"
)

func TestPurgeDeletedAvatar_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		storage := portmocks.NewMockAvatarStorage(ctrl)

		storage.EXPECT().Delete(ctx, "user-1/avatar-1/original").Return(nil)
		storage.EXPECT().Delete(ctx, "user-1/avatar-1/100x100/jpeg").Return(nil)

		uc := NewPurgeDeletedAvatar(storage)
		_, err := uc.Execute(ctx, dto.PurgeDeletedAvatarInput{
			AvatarID: "avatar-1",
			S3Keys:   []string{"user-1/avatar-1/original", "user-1/avatar-1/100x100/jpeg"},
		})
		require.NoError(t, err)
	})

	t.Run("badInput", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		uc := NewPurgeDeletedAvatar(portmocks.NewMockAvatarStorage(ctrl))

		_, err := uc.Execute(ctx, dto.PurgeDeletedAvatarInput{})
		require.ErrorIs(t, err, application.ErrBadInput)
	})
}
