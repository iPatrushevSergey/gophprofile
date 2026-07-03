package usecase

import (
	"context"
	"fmt"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// ProcessUploadedAvatar creates thumbnails and updates processing status.
type ProcessUploadedAvatar struct {
	avatarWriter   appport.AvatarWriter
	avatarStorage  appport.AvatarStorage
	imageProcessor appport.ImageProcessor
	clock          appport.Clock
	tracer         pkgport.Tracer
}

// NewProcessUploadedAvatar returns the process uploaded avatar use case.
func NewProcessUploadedAvatar(
	avatarWriter appport.AvatarWriter,
	avatarStorage appport.AvatarStorage,
	imageProcessor appport.ImageProcessor,
	clock appport.Clock,
	tracer pkgport.Tracer,
) appport.UseCase[dto.ProcessUploadedAvatarInput, struct{}] {
	return &ProcessUploadedAvatar{
		avatarWriter:   avatarWriter,
		avatarStorage:  avatarStorage,
		imageProcessor: imageProcessor,
		clock:          clock,
		tracer:         tracer,
	}
}

// Execute processes uploaded avatar event.
func (uc *ProcessUploadedAvatar) Execute(ctx context.Context, in dto.ProcessUploadedAvatarInput) (_ struct{}, err error) {
	if in.AvatarID == "" || in.UserID == "" || in.S3Key == "" {
		return struct{}{}, application.ErrBadInput
	}

	ctx, span := uc.tracer.Start(ctx, pkgport.SpanConfig{
		Key:  "processing.application.process_uploaded_avatar.execute",
		Name: "process uploaded avatar",
		Kind: pkgport.SpanKindInternal,
		Attributes: []pkgport.Attribute{
			{Key: "avatar_id", Value: in.AvatarID},
			{Key: "user_id", Value: in.UserID},
			{Key: "s3_key", Value: in.S3Key},
		},
	})
	defer func() {
		span.Fail(err)
		span.End()
	}()

	markFailed := func(cause error) (struct{}, error) {
		if err := uc.avatarWriter.UpdateProcessingStatus(ctx, in.AvatarID, vo.ProcessingStatusFailed, uc.clock.Now()); err != nil {
			return struct{}{}, fmt.Errorf("mark processing failed: %w (cause: %w)", err, cause)
		}
		return struct{}{}, cause
	}

	if err = uc.avatarWriter.StartProcessing(ctx, in.AvatarID, uc.clock.Now()); err != nil {
		return struct{}{}, err
	}

	original, err := uc.avatarStorage.Get(ctx, in.S3Key)
	if err != nil {
		return markFailed(err)
	}

	width, height, err := uc.imageProcessor.Dimensions(original)
	if err != nil {
		return markFailed(err)
	}

	sizes := []struct {
		size   vo.ThumbnailSize
		width  int
		height int
	}{
		{vo.ThumbnailSize100, 100, 100},
		{vo.ThumbnailSize300, 300, 300},
		{vo.ThumbnailOriginal, 0, 0},
	}
	formats := []vo.OutputFormat{vo.OutputFormatJPEG, vo.OutputFormatPNG, vo.OutputFormatWebP}

	thumbnailKeys := make(map[vo.ThumbnailSize]map[vo.OutputFormat]string, len(sizes))

	for _, item := range sizes {
		thumb, err := uc.imageProcessor.Resize(original, item.width, item.height)
		if err != nil {
			return markFailed(err)
		}

		formatKeys := make(map[vo.OutputFormat]string, len(formats))
		for _, format := range formats {
			content, err := uc.imageProcessor.Encode(thumb, format)
			if err != nil {
				return markFailed(err)
			}

			key := entity.ThumbnailObjectKey(in.UserID, in.AvatarID, item.size, format)
			if err := uc.avatarStorage.Put(ctx, key, content, format.MimeType()); err != nil {
				return markFailed(err)
			}
			formatKeys[format] = key
		}
		thumbnailKeys[item.size] = formatKeys
	}

	if err = uc.avatarWriter.CompleteProcessing(ctx, dto.CompleteProcessingInput{
		AvatarID:        in.AvatarID,
		ThumbnailS3Keys: thumbnailKeys,
		Width:           width,
		Height:          height,
		UpdatedAt:       uc.clock.Now(),
	}); err != nil {
		return markFailed(err)
	}

	return struct{}{}, nil
}
