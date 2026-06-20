package usecase

import (
	"context"
	"time"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// ExpireUploadingAvatars marks expired upload reservations and removes orphan objects.
type ExpireUploadingAvatars struct {
	avatarReader         appport.AvatarReader
	avatarWriter         appport.AvatarWriter
	avatarStorage        appport.AvatarStorage
	clock                appport.Clock
	uploadReservationTTL time.Duration
}

// NewExpireUploadingAvatars returns the expire uploading avatars use case.
func NewExpireUploadingAvatars(
	avatarReader appport.AvatarReader,
	avatarWriter appport.AvatarWriter,
	avatarStorage appport.AvatarStorage,
	clock appport.Clock,
	uploadReservationTTL time.Duration,
) appport.UseCase[struct{}, struct{}] {
	return &ExpireUploadingAvatars{
		avatarReader:         avatarReader,
		avatarWriter:         avatarWriter,
		avatarStorage:        avatarStorage,
		clock:                clock,
		uploadReservationTTL: uploadReservationTTL,
	}
}

// Execute marks expired upload reservations and removes orphan objects.
func (uc *ExpireUploadingAvatars) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	now := uc.clock.Now()

	avatars, err := uc.avatarReader.ListExpiredUploading(ctx, now.Add(-uc.uploadReservationTTL))
	if err != nil {
		return struct{}{}, err
	}

	// If expired volume grows, switch to chunked batch delete in storage and batch MarkUploadFailed
	// with a ListExpiredUploading limit.
	for _, avatar := range avatars {
		if uc.avatarStorage != nil {
			if err := uc.avatarStorage.Delete(ctx, avatar.S3Key); err != nil {
				return struct{}{}, err
			}
		}
		if err := uc.avatarWriter.MarkUploadFailed(ctx, avatar.ID, now); err != nil {
			return struct{}{}, err
		}
	}

	return struct{}{}, nil
}
