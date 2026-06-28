package dto

import (
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// UploadAvatarInput is application input for avatar upload.
type UploadAvatarInput struct {
	UserID   string
	FileName string
	MimeType string
	Content  []byte
}

// UploadAvatarOutput is application output for avatar upload.
type UploadAvatarOutput struct {
	ID               string
	UserID           string
	UploadStatus     vo.UploadStatus
	ProcessingStatus vo.ProcessingStatus
	CreatedAt        time.Time
}

// GetAvatarInput is application input for fetching avatar.
type GetAvatarInput struct {
	AvatarID      string
	ThumbnailSize vo.ThumbnailSize
	OutputFormat  vo.OutputFormat
}

// GetAvatarOutput is application output for fetching avatar.
type GetAvatarOutput struct {
	AvatarID  string
	Content   []byte
	MimeType  string
	UpdatedAt time.Time
}

// GetAvatarMetadataInput is application input for avatar metadata.
type GetAvatarMetadataInput struct {
	AvatarID string
}

// AvatarMetadataOutput is application output with avatar metadata.
type AvatarMetadataOutput struct {
	ID               string
	UserID           string
	FileName         string
	MimeType         string
	SizeBytes        int64
	Width            int
	Height           int
	ThumbnailS3Keys  map[vo.ThumbnailSize]map[vo.OutputFormat]string
	UploadStatus     vo.UploadStatus
	ProcessingStatus vo.ProcessingStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ListUserAvatarsInput is application input for listing user avatars.
type ListUserAvatarsInput struct {
	UserID string
}

// ListUserAvatarsOutput is application output for listing user avatars.
type ListUserAvatarsOutput struct {
	Items []AvatarMetadataOutput
}

// DeleteAvatarInput is application input for avatar deletion.
type DeleteAvatarInput struct {
	AvatarID      string
	RequestUserID string
}
