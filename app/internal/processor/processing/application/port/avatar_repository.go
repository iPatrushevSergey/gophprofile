package port

import (
	"context"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// AvatarReader provides read access to avatars for processing.
type AvatarReader interface {
	FindByID(ctx context.Context, id string) (*entity.Avatar, error)
}

// AvatarWriter provides write access to avatar processing state.
type AvatarWriter interface {
	StartProcessing(ctx context.Context, id string, updatedAt time.Time) error
	UpdateProcessingStatus(ctx context.Context, id string, status vo.ProcessingStatus, updatedAt time.Time) error
	CompleteProcessing(ctx context.Context, in dto.CompleteProcessingInput) error
}

// AvatarRepo combines read and write access needed by processor.
type AvatarRepo interface {
	AvatarReader
	AvatarWriter
}
