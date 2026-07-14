// Package worker runs background tasks.
package worker

import (
	"context"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

// OutboxPublisherWorker runs a background task to publish pending outbox events.
type OutboxPublisherWorker struct {
	useCases presport.AvatarUseCases
	log      pkgport.Logger
	interval time.Duration
}

// NewOutboxPublisherWorker creates a background task to publish pending outbox events.
func NewOutboxPublisherWorker(
	useCases presport.AvatarUseCases,
	log pkgport.Logger,
	interval time.Duration,
) *OutboxPublisherWorker {
	return &OutboxPublisherWorker{
		useCases: useCases,
		log:      log,
		interval: interval,
	}
}

// Run executes the worker loop.
func (w *OutboxPublisherWorker) Run(ctx context.Context) {
	uc := w.useCases.PublishPendingOutboxEventsUseCase()

	w.log.Info("outbox publisher worker started", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("outbox publisher worker stopped")
			return
		case <-ticker.C:
			if _, err := uc.Execute(ctx, struct{}{}); err != nil {
				w.log.Error("publish pending outbox events failed", "error", err)
			}
		}
	}
}
