package vo

// ThumbnailSize is a supported avatar variant size.
type ThumbnailSize string

const (
	ThumbnailSize100  ThumbnailSize = "100x100"
	ThumbnailSize300  ThumbnailSize = "300x300"
	ThumbnailOriginal ThumbnailSize = "original"
)

// Valid reports whether the thumbnail size is known.
func (s ThumbnailSize) Valid() bool {
	switch s {
	case ThumbnailSize100, ThumbnailSize300, ThumbnailOriginal:
		return true
	default:
		return false
	}
}
