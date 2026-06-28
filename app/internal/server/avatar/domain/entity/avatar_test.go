package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestAvatar_AllS3Keys(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	avatar := NewAvatar(
		"avatar-1",
		"user-1",
		"avatar.png",
		"image/png",
		5,
		"user-1/avatar-1/original",
		vo.UploadStatusCompleted,
		now,
	)
	avatar.ThumbnailS3Keys = map[vo.ThumbnailSize]map[vo.OutputFormat]string{
		vo.ThumbnailSize100: {vo.OutputFormatJPEG: "user-1/avatar-1/100x100"},
		vo.ThumbnailSize300: {vo.OutputFormatJPEG: "user-1/avatar-1/300x300"},
	}

	keys := avatar.AllS3Keys()

	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "user-1/avatar-1/original")
	assert.Contains(t, keys, "user-1/avatar-1/100x100")
	assert.Contains(t, keys, "user-1/avatar-1/300x300")
}
