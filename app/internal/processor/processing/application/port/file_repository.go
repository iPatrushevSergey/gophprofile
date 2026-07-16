package port

import (
	"context"
	"time"
)

// FileRepo abstracts local filesystem operations needed by the processing module.
type FileRepo interface {
	MarkAlive(ctx context.Context, path string, updatedAt time.Time) error
}
