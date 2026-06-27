package imaging

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

var _ appport.ImageResizer = (*Resizer)(nil)

// Resizer creates thumbnail variants from original image bytes.
type Resizer struct{}

// NewResizer creates an image resizer adapter.
func NewResizer() *Resizer {
	return &Resizer{}
}

// Resize resizes image data to target dimensions.
func (r *Resizer) Resize(_ context.Context, data []byte, width, height int) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid resize dimensions: %dx%d", width, height)
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	dst := imaging.Fit(src, width, height, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, dst, imaging.JPEG); err != nil {
		return nil, fmt.Errorf("encode image: %w", err)
	}
	return buf.Bytes(), nil
}
