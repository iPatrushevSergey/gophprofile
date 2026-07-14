package port

import (
	"image"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// ImageProcessor provides image dimension, resize and format conversion operations.
type ImageProcessor interface {
	Dimensions(data []byte) (width, height int, err error)
	Resize(data []byte, width, height int) (image.Image, error)
	Encode(src image.Image, format vo.OutputFormat) ([]byte, error)
}
