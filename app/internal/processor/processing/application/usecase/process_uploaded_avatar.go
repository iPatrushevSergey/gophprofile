package usecase

import (
	"context"
	"fmt"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

const thumbnailContentType = "image/jpeg"

// ProcessUploadedAvatar creates thumbnails and updates processing status.
type ProcessUploadedAvatar struct {
	avatarReader  appport.AvatarReader
	avatarWriter  appport.AvatarWriter
	avatarStorage appport.AvatarStorage
	imageResizer  appport.ImageResizer
	clock         appport.Clock
}

// NewProcessUploadedAvatar returns the process uploaded avatar use case.
func NewProcessUploadedAvatar(
	avatarReader appport.AvatarReader,
	avatarWriter appport.AvatarWriter,
	avatarStorage appport.AvatarStorage,
	imageResizer appport.ImageResizer,
	clock appport.Clock,
) appport.UseCase[dto.ProcessUploadedAvatarInput, struct{}] {
	return &ProcessUploadedAvatar{
		avatarReader:  avatarReader,
		avatarWriter:  avatarWriter,
		avatarStorage: avatarStorage,
		imageResizer:  imageResizer,
		clock:         clock,
	}
}

// Execute processes uploaded avatar event.
func (uc *ProcessUploadedAvatar) Execute(ctx context.Context, in dto.ProcessUploadedAvatarInput) (struct{}, error) {
	if in.AvatarID == "" || in.UserID == "" || in.S3Key == "" {
		return struct{}{}, application.ErrBadInput
	}

	avatar, err := uc.avatarReader.FindByID(ctx, in.AvatarID)
	if err != nil {
		return struct{}{}, err
	}

	if avatar.ProcessingStatus == vo.ProcessingStatusCompleted {
		return struct{}{}, application.ErrAlreadyProcessed
	}

	if err := uc.avatarWriter.UpdateProcessingStatus(ctx, in.AvatarID, vo.ProcessingStatusProcessing, uc.clock.Now()); err != nil {
		return struct{}{}, err
	}

	fail := func(cause error) (struct{}, error) {
		if err := uc.avatarWriter.UpdateProcessingStatus(ctx, in.AvatarID, vo.ProcessingStatusFailed, uc.clock.Now()); err != nil {
			return struct{}{}, fmt.Errorf("mark processing failed: %w (cause: %w)", err, cause)
		}
		return struct{}{}, cause
	}

	original, err := uc.avatarStorage.Get(ctx, in.S3Key)
	if err != nil {
		return fail(err)
	}

	width, height, err := uc.imageResizer.Dimensions(original)
	if err != nil {
		return fail(err)
	}

	sizes := []struct {
		size   vo.ThumbnailSize
		width  int
		height int
	}{
		{vo.ThumbnailSize100, 100, 100},
		{vo.ThumbnailSize300, 300, 300},
	}

	thumbnailKeys := make(map[vo.ThumbnailSize]string, len(sizes))

	for _, item := range sizes {
		resized, err := uc.imageResizer.Resize(ctx, original, item.width, item.height)
		if err != nil {
			return fail(err)
		}

		key := entity.ThumbnailObjectKey(in.UserID, in.AvatarID, item.size)
		if err := uc.avatarStorage.Put(ctx, key, resized, thumbnailContentType); err != nil {
			return fail(err)
		}
		thumbnailKeys[item.size] = key
	}

	if err := uc.avatarWriter.CompleteProcessing(ctx, dto.CompleteProcessingInput{
		AvatarID:        in.AvatarID,
		ThumbnailS3Keys: thumbnailKeys,
		Width:           width,
		Height:          height,
		UpdatedAt:       uc.clock.Now(),
	}); err != nil {
		return fail(err)
	}

	return struct{}{}, nil
}
