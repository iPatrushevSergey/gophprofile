package usecase

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/imaging"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

func minimalPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestProcessUploadedAvatar_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		clock := portmocks.NewMockClock(ctrl)

		pngData := minimalPNG(t)
		reader.EXPECT().FindByID(ctx, "avatar-1").Return(&entity.Avatar{
			ProcessingStatus: vo.ProcessingStatusPending,
		}, nil)
		clock.EXPECT().Now().Return(now).AnyTimes()
		writer.EXPECT().UpdateProcessingStatus(ctx, "avatar-1", vo.ProcessingStatusProcessing, now).Return(nil)
		storage.EXPECT().Get(ctx, "user-1/avatar-1/original").Return(pngData, nil)
		storage.EXPECT().Put(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(9)
		writer.EXPECT().CompleteProcessing(ctx, gomock.Any()).Return(nil)

		uc := NewProcessUploadedAvatar(reader, writer, storage, imaging.NewProcessor(), clock)
		_, err := uc.Execute(ctx, dto.ProcessUploadedAvatarInput{
			AvatarID: "avatar-1",
			UserID:   "user-1",
			S3Key:    "user-1/avatar-1/original",
		})
		require.NoError(t, err)
	})

	t.Run("badInput", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		uc := NewProcessUploadedAvatar(
			portmocks.NewMockAvatarReader(ctrl),
			portmocks.NewMockAvatarWriter(ctrl),
			portmocks.NewMockAvatarStorage(ctrl),
			imaging.NewProcessor(),
			portmocks.NewMockClock(ctrl),
		)

		_, err := uc.Execute(ctx, dto.ProcessUploadedAvatarInput{})
		require.ErrorIs(t, err, application.ErrBadInput)
	})

	t.Run("alreadyProcessed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		reader.EXPECT().FindByID(ctx, "avatar-1").Return(&entity.Avatar{
			ProcessingStatus: vo.ProcessingStatusCompleted,
		}, nil)

		uc := NewProcessUploadedAvatar(
			reader,
			portmocks.NewMockAvatarWriter(ctrl),
			portmocks.NewMockAvatarStorage(ctrl),
			imaging.NewProcessor(),
			portmocks.NewMockClock(ctrl),
		)

		_, err := uc.Execute(ctx, dto.ProcessUploadedAvatarInput{
			AvatarID: "avatar-1",
			UserID:   "user-1",
			S3Key:    "user-1/avatar-1/original",
		})
		require.ErrorIs(t, err, application.ErrAlreadyProcessed)
	})
}
