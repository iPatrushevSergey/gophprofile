package vo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestThumbnailSize_Valid(t *testing.T) {
	assert.True(t, vo.ThumbnailSize100.Valid())
	assert.True(t, vo.ThumbnailOriginal.Valid())
	assert.False(t, vo.ThumbnailSize("999x999").Valid())
}

func TestOutputFormat_ValidAndMimeType(t *testing.T) {
	assert.True(t, vo.OutputFormatJPEG.Valid())
	assert.True(t, vo.OutputFormat("").Valid())
	assert.False(t, vo.OutputFormat("gif").Valid())

	assert.Equal(t, "image/jpeg", vo.OutputFormatJPEG.MimeType())
	assert.Equal(t, "image/png", vo.OutputFormatPNG.MimeType())
	assert.Equal(t, "image/webp", vo.OutputFormatWebP.MimeType())
	assert.Equal(t, "", vo.OutputFormat("gif").MimeType())
}

func TestUploadStatus_Valid(t *testing.T) {
	assert.True(t, vo.UploadStatusCompleted.Valid())
	assert.False(t, vo.UploadStatus("unknown").Valid())
}

func TestProcessingStatus_Valid(t *testing.T) {
	assert.True(t, vo.ProcessingStatusPending.Valid())
	assert.False(t, vo.ProcessingStatus("unknown").Valid())
}
