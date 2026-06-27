package model

//go:generate go run github.com/mailru/easyjson/easyjson@v0.9.0 -all $GOFILE

// AvatarUploadedEvent is the RabbitMQ wire payload for avatar uploaded events.
type AvatarUploadedEvent struct {
	AvatarID string `json:"avatar_id"`
	UserID   string `json:"user_id"`
	S3Key    string `json:"s3_key"`
}

// AvatarDeletedEvent is the RabbitMQ wire payload for avatar deleted events.
type AvatarDeletedEvent struct {
	AvatarID string   `json:"avatar_id"`
	S3Keys   []string `json:"s3_keys"`
}
