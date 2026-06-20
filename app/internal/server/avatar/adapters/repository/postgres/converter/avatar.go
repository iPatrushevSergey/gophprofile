package converter

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen .

// goverter:converter
// goverter:output:file generated/avatar.go
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:extend UUIDToString
// goverter:extend StringToUUID
// goverter:extend StringToUploadStatus
// goverter:extend UploadStatusToString
// goverter:extend StringToProcessingStatus
// goverter:extend ProcessingStatusToString
// goverter:extend RawMessageToThumbnailS3Keys
// goverter:extend ThumbnailS3KeysToRawMessage
type AvatarConverter interface {
	AvatarModelToAvatarEntity(source model.Avatar) (entity.Avatar, error)
	AvatarEntityToAvatarModel(source entity.Avatar) (model.Avatar, error)
}

// CopyTime maps time.Time for goverter (required: time.Time has unexported fields).
func CopyTime(v time.Time) time.Time {
	return v
}

// CopyTimePtr maps *time.Time for goverter (required: time.Time has unexported fields).
func CopyTimePtr(v *time.Time) *time.Time {
	return v
}

// UUIDToString converts a database UUID to a domain avatar id.
func UUIDToString(id uuid.UUID) string {
	return id.String()
}

// StringToUUID converts a domain avatar id to a database UUID.
func StringToUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// StringToUploadStatus parses upload status from database text.
func StringToUploadStatus(raw string) vo.UploadStatus {
	return vo.UploadStatus(raw)
}

// UploadStatusToString converts upload status to database text.
func UploadStatusToString(status vo.UploadStatus) string {
	return string(status)
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

// ThumbnailS3KeysToRawMessage converts thumbnail keys map to JSONB.
func ThumbnailS3KeysToRawMessage(keys map[vo.ThumbnailSize]string) (json.RawMessage, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	raw, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
