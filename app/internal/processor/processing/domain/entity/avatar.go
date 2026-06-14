package entity

import (
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// Avatar is a read/write view of avatar row used by processor.
type Avatar struct {
	ID               string
	UserID           string
	S3Key            string
	ThumbnailS3Keys  map[vo.ThumbnailSize]string
	ProcessingStatus vo.ProcessingStatus
	UpdatedAt        time.Time
}
