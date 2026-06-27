package dto

import (
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// ProcessUploadedAvatarInput is application input for processing uploaded avatar.
type ProcessUploadedAvatarInput struct {
	AvatarID string
	UserID   string
	S3Key    string
}

// PurgeDeletedAvatarInput is application input for deleting avatar files from object storage.
type PurgeDeletedAvatarInput struct {
	AvatarID string
	S3Keys   []string
}

// CompleteProcessingInput is application input for completing avatar processing.
type CompleteProcessingInput struct {
	AvatarID        string
	ThumbnailS3Keys map[vo.ThumbnailSize]string
	UpdatedAt       time.Time
}
