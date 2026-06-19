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
	avatar.ThumbnailS3Keys[vo.ThumbnailSize100] = "user-1/avatar-1/100x100"
	avatar.ThumbnailS3Keys[vo.ThumbnailSize300] = "user-1/avatar-1/300x300"

	keys := avatar.AllS3Keys()

	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "user-1/avatar-1/original")
	assert.Contains(t, keys, "user-1/avatar-1/100x100")
	assert.Contains(t, keys, "user-1/avatar-1/300x300")
	assert.Equal(t, "user-1/avatar-1/original", keys[0])
}
