package vo

// OutputFormat is a requested image encoding for avatar download.
type OutputFormat string

const (
	OutputFormatJPEG OutputFormat = "jpeg"
	OutputFormatPNG  OutputFormat = "png"
	OutputFormatWebP OutputFormat = "webp"
)

// Valid reports whether format is supported for ?format= query.
func (f OutputFormat) Valid() bool {
	switch f {
	case "", OutputFormatJPEG, OutputFormatPNG, OutputFormatWebP:
		return true
	default:
		return false
	}
}

// MimeType returns HTTP content type for the output format.
func (f OutputFormat) MimeType() string {
	switch f {
	case OutputFormatJPEG:
		return "image/jpeg"
	case OutputFormatPNG:
		return "image/png"
	case OutputFormatWebP:
		return "image/webp"
	}
	return ""
}
