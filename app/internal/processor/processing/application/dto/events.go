package dto

// AvatarUploadedEvent is consumed after a successful avatar upload.
type AvatarUploadedEvent struct {
	AvatarID string
	UserID   string
	S3Key    string
}

// AvatarDeletedEvent is consumed after avatar soft delete.
type AvatarDeletedEvent struct {
	AvatarID string
	S3Keys   []string
}
