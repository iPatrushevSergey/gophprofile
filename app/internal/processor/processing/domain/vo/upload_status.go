package vo

// UploadStatus describes original file upload progress.
type UploadStatus string

const (
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusFailed    UploadStatus = "failed"
)

// Valid reports whether the upload status is known.
func (s UploadStatus) Valid() bool {
	switch s {
	case UploadStatusUploading, UploadStatusCompleted, UploadStatusFailed:
		return true
	default:
		return false
	}
}
