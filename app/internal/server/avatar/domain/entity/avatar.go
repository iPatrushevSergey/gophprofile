package entity

import (
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
