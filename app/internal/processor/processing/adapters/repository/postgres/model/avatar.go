package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Avatar is the DB projection of the avatars table row.
type Avatar struct {
	ID               uuid.UUID
	UserID           string
	FileName         string
	MimeType         string
	SizeBytes        int64
	Width            *int
	Height           *int
	S3Key            string
	ThumbnailS3Keys  json.RawMessage
	UploadStatus     string
	ProcessingStatus string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
