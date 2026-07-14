package port

import "context"

// FileRepo abstracts local filesystem operations needed by the processing module.
type FileRepo interface {
	MarkAlive(ctx context.Context, path string) error
}
