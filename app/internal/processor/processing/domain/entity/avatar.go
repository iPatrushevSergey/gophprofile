package entity

import (
	"fmt"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// Avatar is the projection of the avatar row.
type Avatar struct {
	ID               string
	UserID           string
	S3Key            string
	ThumbnailS3Keys  map[vo.ThumbnailSize]map[vo.OutputFormat]string
	ProcessingStatus vo.ProcessingStatus
	UpdatedAt        time.Time
}

// ThumbnailObjectKey builds object storage key for a thumbnail variant.
func ThumbnailObjectKey(userID, avatarID string, size vo.ThumbnailSize, format vo.OutputFormat) string {
	return fmt.Sprintf("%s/%s/%s/%s", userID, avatarID, size, format)
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
