package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// PurgeDeletedAvatar removes avatar files from object storage.
type PurgeDeletedAvatar struct {
	avatarStorage appport.AvatarStorage
}

// NewPurgeDeletedAvatar returns the purge deleted avatar use case.
func NewPurgeDeletedAvatar(avatarStorage appport.AvatarStorage) appport.UseCase[dto.PurgeDeletedAvatarInput, struct{}] {
	return &PurgeDeletedAvatar{avatarStorage: avatarStorage}
}

// Execute purges deleted avatar files.
func (uc *PurgeDeletedAvatar) Execute(ctx context.Context, in dto.PurgeDeletedAvatarInput) (struct{}, error) {
	if in.AvatarID == "" || len(in.S3Keys) == 0 {
		return struct{}{}, application.ErrBadInput
	}

	for _, key := range in.S3Keys {
		if err := uc.avatarStorage.Delete(ctx, key); err != nil {
			return struct{}{}, err
		}
	}

	return struct{}{}, nil
}
