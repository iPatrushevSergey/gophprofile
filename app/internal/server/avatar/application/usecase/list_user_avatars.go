package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// ListUserAvatars returns all active avatars for a user.
type ListUserAvatars struct {
	avatarReader appport.AvatarReader
}

// NewListUserAvatars returns the list user avatars use case.
func NewListUserAvatars(avatarReader appport.AvatarReader) appport.UseCase[dto.ListUserAvatarsInput, dto.ListUserAvatarsOutput] {
	return &ListUserAvatars{avatarReader: avatarReader}
}

// Execute returns user avatars.
func (uc *ListUserAvatars) Execute(ctx context.Context, in dto.ListUserAvatarsInput) (dto.ListUserAvatarsOutput, error) {
	avatars, err := uc.avatarReader.ListByUserID(ctx, in.UserID)
	if err != nil {
		return dto.ListUserAvatarsOutput{}, err
	}

	items := make([]dto.AvatarMetadataOutput, 0, len(avatars))
	for _, avatar := range avatars {
		items = append(items, dto.AvatarMetadataOutput{
			ID:               avatar.ID,
			UserID:           avatar.UserID,
			FileName:         avatar.FileName,
			MimeType:         avatar.MimeType,
			SizeBytes:        avatar.SizeBytes,
			ThumbnailS3Keys:  avatar.ThumbnailS3Keys,
			UploadStatus:     avatar.UploadStatus,
			ProcessingStatus: avatar.ProcessingStatus,
			CreatedAt:        avatar.CreatedAt,
			UpdatedAt:        avatar.UpdatedAt,
		})
	}

	return dto.ListUserAvatarsOutput{Items: items}, nil
}
