package port

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
)

// AvatarReader provides read access to avatars.
type AvatarReader interface {
	FindByID(ctx context.Context, id string) (*entity.Avatar, error)
	ListByUserID(ctx context.Context, userID string) ([]entity.Avatar, error)
	ListExpiredUploading(ctx context.Context, before time.Time) ([]entity.Avatar, error)
}

// AvatarWriter provides write access to avatars.
type AvatarWriter interface {
	Create(ctx context.Context, avatar *entity.Avatar) error
	MarkUploadCompleted(ctx context.Context, id string, updatedAt time.Time) error
	MarkUploadFailed(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id, userID string) error
}

// AvatarRepo combines read and write access to avatars.
type AvatarRepo interface {
	AvatarReader
	AvatarWriter
}
