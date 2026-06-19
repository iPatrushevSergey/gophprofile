package entity

import (
	"fmt"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// Avatar is the avatar aggregate root.
type Avatar struct {
	ID               string
	UserID           string
	FileName         string
	MimeType         string
	SizeBytes        int64
	S3Key            string
	ThumbnailS3Keys  map[vo.ThumbnailSize]string
	UploadStatus     vo.UploadStatus
	ProcessingStatus vo.ProcessingStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

// NewAvatar creates a new avatar with the given upload status.
func NewAvatar(
	id, userID, fileName, mimeType string,
	sizeBytes int64,
	s3Key string,
	uploadStatus vo.UploadStatus,
	now time.Time,
) *Avatar {
	return &Avatar{
		ID:               id,
		UserID:           userID,
		FileName:         fileName,
		MimeType:         mimeType,
		SizeBytes:        sizeBytes,
		S3Key:            s3Key,
		ThumbnailS3Keys:  make(map[vo.ThumbnailSize]string),
		UploadStatus:     uploadStatus,
		ProcessingStatus: vo.ProcessingStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// OriginalObjectKey builds object storage key for the original image.
func OriginalObjectKey(userID, avatarID string) string {
	return fmt.Sprintf("%s/%s/original", userID, avatarID)
}

// AllS3Keys returns original and thumbnail object storage keys.
func (a Avatar) AllS3Keys() []string {
	keys := make([]string, 0, len(a.ThumbnailS3Keys)+1)
	keys = append(keys, a.S3Key)
	for _, key := range a.ThumbnailS3Keys {
		keys = append(keys, key)
	}
	return keys
}
