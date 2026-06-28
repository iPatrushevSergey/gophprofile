package vo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

func TestThumbnailSize_Valid(t *testing.T) {
	assert.True(t, vo.ThumbnailSize300.Valid())
	assert.False(t, vo.ThumbnailSize("bad").Valid())
}

func TestOutputFormat_ValidAndMimeType(t *testing.T) {
	assert.True(t, vo.OutputFormatWebP.Valid())
	assert.False(t, vo.OutputFormat("bmp").Valid())
	assert.Equal(t, "image/webp", vo.OutputFormatWebP.MimeType())
}

func TestProcessingStatus_Valid(t *testing.T) {
	assert.True(t, vo.ProcessingStatusCompleted.Valid())
	assert.False(t, vo.ProcessingStatus("bad").Valid())
}
