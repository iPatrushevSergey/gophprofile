package dto

//go:generate go run github.com/mailru/easyjson/easyjson@v0.9.0 -all $GOFILE

import "time"

// UploadAvatarRequest is a presentation request for avatar upload.
type UploadAvatarRequest struct {
	FileName string
	MimeType string
	Content  []byte
}

// UploadAvatarResponse is a presentation response for avatar upload.
type UploadAvatarResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AvatarMetadataResponse is a presentation response with avatar metadata.
type AvatarMetadataResponse struct {
	ID         string           `json:"id"`
	UserID     string           `json:"user_id"`
	FileName   string           `json:"file_name"`
	MimeType   string           `json:"mime_type"`
	SizeBytes  int64            `json:"size_bytes"`
	Dimensions *ImageDimensions `json:"dimensions,omitempty"`
	Thumbnails []ThumbnailURL   `json:"thumbnails"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// ImageDimensions describes original image size in pixels.
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// UserAvatarsResponse is a presentation response with user avatar list.
type UserAvatarsResponse struct {
	Items []AvatarMetadataResponse `json:"items"`
}

// ThumbnailURL describes one thumbnail variant URL.
type ThumbnailURL struct {
	Size string `json:"size"`
	URL  string `json:"url"`
}

// ErrorResponse is a common API error payload.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	MaxSize int64  `json:"max_size,omitempty"`
}

// HealthResponse is a presentation response for service health check.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Storage  string `json:"storage"`
	Broker   string `json:"broker"`
}
