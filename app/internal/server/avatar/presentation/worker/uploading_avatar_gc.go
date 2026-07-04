// Package worker runs background tasks.
package worker

import (
	"context"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

// UploadingAvatarGCWorker runs a background task to expire uploading avatar reservations.
type UploadingAvatarGCWorker struct {
	useCases presport.AvatarUseCases
	log      pkgport.Logger
	interval time.Duration
}

// NewUploadingAvatarGCWorker creates a background task to expire uploading avatar reservations.
func NewUploadingAvatarGCWorker(
	useCases presport.AvatarUseCases,
	log pkgport.Logger,
	interval time.Duration,
) *UploadingAvatarGCWorker {
	return &UploadingAvatarGCWorker{
		useCases: useCases,
		log:      log,
		interval: interval,
	}
}

// Run executes the worker loop.
func (w *UploadingAvatarGCWorker) Run(ctx context.Context) {
	uc := w.useCases.ExpireUploadingAvatarsUseCase()

	w.log.Info(ctx, "uploading avatar gc worker started", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info(ctx, "uploading avatar gc worker stopped")
			return
		case <-ticker.C:
			if _, err := uc.Execute(ctx, struct{}{}); err != nil {
				w.log.Error(ctx, "expire uploading avatars failed", "error", err)
			}
		}
	}
}
