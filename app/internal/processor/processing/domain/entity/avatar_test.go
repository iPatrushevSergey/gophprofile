package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

func TestAvatar_AllS3Keys(t *testing.T) {
	avatar := Avatar{
		S3Key: "user-1/avatar-1/original",
		ThumbnailS3Keys: map[vo.ThumbnailSize]map[vo.OutputFormat]string{
			vo.ThumbnailSize100: {vo.OutputFormatJPEG: "user-1/avatar-1/100x100/jpeg"},
		},
	}

	keys := avatar.AllS3Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "user-1/avatar-1/original")
}

func TestThumbnailObjectKey(t *testing.T) {
	key := ThumbnailObjectKey("user-1", "avatar-1", vo.ThumbnailSize300, vo.OutputFormatWebP)
	assert.Equal(t, "user-1/avatar-1/300x300/webp", key)
}
