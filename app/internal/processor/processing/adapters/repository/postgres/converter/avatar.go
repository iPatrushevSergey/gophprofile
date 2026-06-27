package converter

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen .

// goverter:converter
// goverter:output:file generated/avatar.go
// goverter:extend CopyTime
// goverter:extend UUIDToString
// goverter:extend StringToProcessingStatus
// goverter:extend ProcessingStatusToString
// goverter:extend RawMessageToThumbnailS3Keys
type AvatarConverter interface {
	AvatarModelToAvatarEntity(source model.Avatar) (entity.Avatar, error)
}

// CopyTime maps time.Time for goverter (required: time.Time has unexported fields).
func CopyTime(v time.Time) time.Time {
	return v
}

// UUIDToString converts a database UUID to a domain avatar id.
func UUIDToString(id uuid.UUID) string {
	return id.String()
}

// StringToProcessingStatus parses processing status from database text.
func StringToProcessingStatus(raw string) vo.ProcessingStatus {
	return vo.ProcessingStatus(raw)
}

// ProcessingStatusToString converts processing status to database text.
func ProcessingStatusToString(status vo.ProcessingStatus) string {
	return string(status)
}

// RawMessageToThumbnailS3Keys converts JSONB to thumbnail keys map.
func RawMessageToThumbnailS3Keys(raw json.RawMessage) (map[vo.ThumbnailSize]string, error) {
	if len(raw) == 0 {
		return make(map[vo.ThumbnailSize]string), nil
	}

	keys := make(map[vo.ThumbnailSize]string)
	if err := json.Unmarshal(raw, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}
