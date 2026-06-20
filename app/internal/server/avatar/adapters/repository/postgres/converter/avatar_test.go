package converter

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestAvatarConverter_roundTrip(t *testing.T) {
	c := AvatarConverterImpl{}
	now := time.Now().UTC().Truncate(time.Second)
	avatar := entity.Avatar{
		ID:        uuid.NewString(),
		UserID:    "user-1",
		FileName:  "avatar.png",
		MimeType:  "image/png",
		SizeBytes: 5,
		S3Key:     "user-1/avatar-1/original",
		ThumbnailS3Keys: map[vo.ThumbnailSize]string{
			vo.ThumbnailSize100: "user-1/avatar-1/100x100",
		},
		UploadStatus:     vo.UploadStatusCompleted,
		ProcessingStatus: vo.ProcessingStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	row, err := c.ToModel(avatar)
	require.NoError(t, err)

	got, err := c.ToEntity(row)
	require.NoError(t, err)
	assert.Equal(t, avatar, got)
}
