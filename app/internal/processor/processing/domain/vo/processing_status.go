package vo

// ProcessingStatus describes async thumbnail processing progress.
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// Valid reports whether the processing status is known.
func (s ProcessingStatus) Valid() bool {
	switch s {
	case ProcessingStatusPending, ProcessingStatusProcessing, ProcessingStatusCompleted, ProcessingStatusFailed:
		return true
	default:
		return false
	}
}
