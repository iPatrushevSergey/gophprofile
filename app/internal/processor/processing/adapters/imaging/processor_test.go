package imaging

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

func minimalPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestProcessor_DimensionsResizeEncode(t *testing.T) {
	p := NewProcessor()
	data := minimalPNG(t)

	width, height, err := p.Dimensions(data)
	require.NoError(t, err)
	assert.Equal(t, 8, width)
	assert.Equal(t, 8, height)

	resized, err := p.Resize(data, 4, 4)
	require.NoError(t, err)
	assert.Equal(t, 4, resized.Bounds().Dx())

	original, err := p.Resize(data, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 8, original.Bounds().Dx())

	jpeg, err := p.Encode(resized, vo.OutputFormatJPEG)
	require.NoError(t, err)
	assert.NotEmpty(t, jpeg)

	_, err = p.Encode(resized, vo.OutputFormat("gif"))
	assert.Error(t, err)
}
