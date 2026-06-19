package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// GetAvatar returns an avatar by id and requested size.
type GetAvatar struct {
	avatarReader  appport.AvatarReader
	avatarStorage appport.AvatarStorage
}

// NewGetAvatar returns the get avatar use case.
func NewGetAvatar(
	avatarReader appport.AvatarReader,
	avatarStorage appport.AvatarStorage,
) appport.UseCase[dto.GetAvatarInput, dto.GetAvatarOutput] {
	return &GetAvatar{
		avatarReader:  avatarReader,
		avatarStorage: avatarStorage,
	}
}

// Execute returns an avatar by id and requested size.

func (uc *GetAvatar) Execute(ctx context.Context, in dto.GetAvatarInput) (dto.GetAvatarOutput, error) {
	size := in.ThumbnailSize
	avatar, err := uc.avatarReader.FindByID(ctx, in.AvatarID)
	if err != nil {
		return dto.GetAvatarOutput{}, err
	}

	var key string
	switch size {
	case vo.ThumbnailOriginal:
		key = avatar.S3Key
	case vo.ThumbnailSize100, vo.ThumbnailSize300:
		var ok bool
		key, ok = avatar.ThumbnailS3Keys[size]
		if !ok {
			return dto.GetAvatarOutput{}, application.ErrNotFound
		}
	}

	content, err := uc.avatarStorage.Get(ctx, key)
	if err != nil {
		return dto.GetAvatarOutput{}, err
	}

	return dto.GetAvatarOutput{
		AvatarID:  avatar.ID,
		Content:   content,
		MimeType:  avatar.MimeType,
		UpdatedAt: avatar.UpdatedAt,
	}, nil
}
