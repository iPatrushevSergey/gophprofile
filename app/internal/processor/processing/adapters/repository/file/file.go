// Package file implements filesystem-backed repositories for the processing module.
package file

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// FileRepository performs local filesystem operations.
type FileRepository struct{}

// NewFileRepository creates a filesystem repository.
func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

// MarkAlive creates or updates the file mtime at path.
func (r *FileRepository) MarkAlive(_ context.Context, path string, updatedAt time.Time) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("file path is required")
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close file: %w", err)
	}
	if err := os.Chtimes(path, updatedAt, updatedAt); err != nil {
		return fmt.Errorf("update file mtime: %w", err)
	}
	return nil
}
