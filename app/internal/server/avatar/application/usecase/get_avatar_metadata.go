package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// GetAvatarMetadata returns avatar metadata by id.
type GetAvatarMetadata struct {
	avatarReader appport.AvatarReader
}

// NewGetAvatarMetadata returns the get avatar metadata use case.
func NewGetAvatarMetadata(avatarReader appport.AvatarReader) appport.UseCase[dto.GetAvatarMetadataInput, dto.AvatarMetadataOutput] {
	return &GetAvatarMetadata{avatarReader: avatarReader}
}

// Execute returns avatar metadata.
func (uc *GetAvatarMetadata) Execute(ctx context.Context, in dto.GetAvatarMetadataInput) (dto.AvatarMetadataOutput, error) {
	avatar, err := uc.avatarReader.FindByID(ctx, in.AvatarID)
	if err != nil {
		return dto.AvatarMetadataOutput{}, err
	}

	return dto.AvatarMetadataOutput{
		ID:               avatar.ID,
		UserID:           avatar.UserID,
		FileName:         avatar.FileName,
		MimeType:         avatar.MimeType,
		SizeBytes:        avatar.SizeBytes,
		Width:            avatar.Width,
		Height:           avatar.Height,
		ThumbnailS3Keys:  avatar.ThumbnailS3Keys,
		UploadStatus:     avatar.UploadStatus,
		ProcessingStatus: avatar.ProcessingStatus,
		CreatedAt:        avatar.CreatedAt,
		UpdatedAt:        avatar.UpdatedAt,
	}, nil
}
