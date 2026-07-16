// Package worker runs background tasks.
package worker

import (
	"context"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// HealthFileWorker runs a background task that refreshes the processor health file.
type HealthFileWorker struct {
	refreshHealthFile appport.UseCase[struct{}, struct{}]
	log               pkgport.Logger
	interval          time.Duration
}

// NewHealthFileWorker creates a background task that refreshes the health file.
func NewHealthFileWorker(
	refreshHealthFile appport.UseCase[struct{}, struct{}],
	log pkgport.Logger,
	interval time.Duration,
) *HealthFileWorker {
	return &HealthFileWorker{
		refreshHealthFile: refreshHealthFile,
		log:               log,
		interval:          interval,
	}
}

// Run executes the worker loop.
func (w *HealthFileWorker) Run(ctx context.Context) {
	w.log.Info(ctx, "health file worker started", "interval", w.interval)

	if _, err := w.refreshHealthFile.Execute(ctx, struct{}{}); err != nil {
		w.log.Warn(ctx, "refresh health file failed", "error", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info(ctx, "health file worker stopped")
			return
		case <-ticker.C:
			workCtx := context.WithoutCancel(ctx)
			if _, err := w.refreshHealthFile.Execute(workCtx, struct{}{}); err != nil {
				w.log.Warn(workCtx, "refresh health file failed", "error", err)
			}
		}
	}
}
