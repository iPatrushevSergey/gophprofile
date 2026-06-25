package port

import "context"

// ImageResizer creates thumbnail variants from original image bytes.
type ImageResizer interface {
	Resize(ctx context.Context, data []byte, width, height int) ([]byte, error)
}
