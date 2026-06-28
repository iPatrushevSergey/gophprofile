package imaging

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/KarpelesLab/gowebp"
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

var _ appport.ImageProcessor = (*Processor)(nil)

// Processor handles raster image dimensions, resize and format conversion.
type Processor struct{}

// NewProcessor creates an image processor adapter.
func NewProcessor() *Processor {
	return &Processor{}
}

// Dimensions returns image width and height without full decode.
func (p *Processor) Dimensions(data []byte) (int, int, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("read image dimensions: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

// Resize decodes image bytes and fits the image into target dimensions.
// Pass zero width and height to decode without resizing.
func (p *Processor) Resize(data []byte, width, height int) (image.Image, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	if width <= 0 || height <= 0 {
		return src, nil
	}
	return imaging.Fit(src, width, height, imaging.Lanczos), nil
}

// Encode converts the image to the requested output format.
func (p *Processor) Encode(src image.Image, format vo.OutputFormat) ([]byte, error) {
	var buf bytes.Buffer

	switch format {
	case vo.OutputFormatJPEG:
		if err := imaging.Encode(&buf, src, imaging.JPEG); err != nil {
			return nil, fmt.Errorf("encode jpeg: %w", err)
		}
	case vo.OutputFormatPNG:
		if err := imaging.Encode(&buf, src, imaging.PNG); err != nil {
			return nil, fmt.Errorf("encode png: %w", err)
		}
	case vo.OutputFormatWebP:
		if err := gowebp.Encode(&buf, src, &gowebp.Options{Lossy: true, Quality: 80}); err != nil {
			return nil, fmt.Errorf("encode webp: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}

	return buf.Bytes(), nil
}
