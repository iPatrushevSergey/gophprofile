package port

import "context"

// ImageResizer reads image dimensions and creates thumbnail variants.
type ImageResizer interface {
	Dimensions(data []byte) (width, height int, err error)
	Resize(ctx context.Context, data []byte, width, height int) ([]byte, error)
}
