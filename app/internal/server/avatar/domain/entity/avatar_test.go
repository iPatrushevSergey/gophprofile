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

func TestOriginalObjectKey(t *testing.T) {
	assert.Equal(t, "user-1/avatar-1/original", OriginalObjectKey("user-1", "avatar-1"))
}

func TestAvatar_LookupVariantKey(t *testing.T) {
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
		vo.ThumbnailSize100: {vo.OutputFormatJPEG: "user-1/avatar-1/100x100/jpeg"},
	}

	key, mime, ok := avatar.LookupVariantKey(vo.ThumbnailOriginal, "")
	assert.True(t, ok)
	assert.Equal(t, "user-1/avatar-1/original", key)
	assert.Equal(t, "image/png", mime)

	key, mime, ok = avatar.LookupVariantKey(vo.ThumbnailSize100, vo.OutputFormatJPEG)
	assert.True(t, ok)
	assert.Equal(t, "user-1/avatar-1/100x100/jpeg", key)
	assert.Equal(t, "image/jpeg", mime)

	_, _, ok = avatar.LookupVariantKey(vo.ThumbnailSize300, vo.OutputFormatJPEG)
	assert.False(t, ok)

	_, _, ok = avatar.LookupVariantKey(vo.ThumbnailSize100, vo.OutputFormat("gif"))
	assert.False(t, ok)
}
