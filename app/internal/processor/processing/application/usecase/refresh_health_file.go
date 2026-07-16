package usecase

import (
	"context"
	"fmt"
	"strings"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// RefreshHealthFile refreshes a local health marker file.
type RefreshHealthFile struct {
	fileRepo appport.FileRepo
	clock    appport.Clock
	path     string
}

// NewRefreshHealthFile returns the refresh health file use case.
func NewRefreshHealthFile(
	fileRepo appport.FileRepo,
	clock appport.Clock,
	path string,
) appport.UseCase[struct{}, struct{}] {
	return &RefreshHealthFile{
		fileRepo: fileRepo,
		clock:    clock,
		path:     strings.TrimSpace(path),
	}
}

// Execute refreshes the health marker file.
func (uc *RefreshHealthFile) Execute(ctx context.Context, _ struct{}) (struct{}, error) {
	if uc.path == "" {
		return struct{}{}, fmt.Errorf("health file path is required")
	}
	if err := uc.fileRepo.MarkAlive(ctx, uc.path, uc.clock.Now()); err != nil {
		return struct{}{}, fmt.Errorf("mark health file alive: %w", err)
	}
	return struct{}{}, nil
}
