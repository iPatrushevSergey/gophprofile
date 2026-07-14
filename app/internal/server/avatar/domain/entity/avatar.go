package entity

import (
	"fmt"
	"strings"
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
	Width            int
	Height           int
	S3Key            string
	ThumbnailS3Keys  map[vo.ThumbnailSize]map[vo.OutputFormat]string
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
		ThumbnailS3Keys:  make(map[vo.ThumbnailSize]map[vo.OutputFormat]string),
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
	keys := make([]string, 0, len(a.ThumbnailS3Keys)*3+1)
	keys = append(keys, a.S3Key)
	for _, formatKeys := range a.ThumbnailS3Keys {
		for _, key := range formatKeys {
			keys = append(keys, key)
		}
	}
	return keys
}

// LookupVariantKey returns storage key and MIME type for the requested variant.
func (a Avatar) LookupVariantKey(size vo.ThumbnailSize, format vo.OutputFormat) (string, string, bool) {
	if format != "" && !format.Valid() {
		return "", "", false
	}

	if size == vo.ThumbnailOriginal && (format == "" || strings.EqualFold(format.MimeType(), a.MimeType)) {
		return a.S3Key, a.MimeType, true
	}

	if format == "" {
		format = vo.OutputFormatJPEG
	}

	formatKeys, ok := a.ThumbnailS3Keys[size]
	if !ok {
		return "", "", false
	}

	key, ok := formatKeys[format]
	if !ok {
		return "", "", false
	}

	return key, format.MimeType(), true
}
