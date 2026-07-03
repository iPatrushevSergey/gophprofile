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
	tracer   pkgport.Tracer
	interval time.Duration
}

// NewOutboxPublisherWorker creates a background task to publish pending outbox events.
func NewOutboxPublisherWorker(
	useCases presport.AvatarUseCases,
	log pkgport.Logger,
	tracer pkgport.Tracer,
	interval time.Duration,
) *OutboxPublisherWorker {
	return &OutboxPublisherWorker{
		useCases: useCases,
		log:      log,
		tracer:   tracer,
		interval: interval,
	}
}

// Run executes the worker loop.
func (w *OutboxPublisherWorker) Run(ctx context.Context) {
	w.log.Info("outbox publisher worker started", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("outbox publisher worker stopped")
			return
		case <-ticker.C:
			tickCtx, span := w.tracer.Start(ctx, pkgport.SpanConfig{
				Key:  "avatar.presentation.outbox_publisher.publish_pending_outbox_events",
				Name: "publish pending outbox events",
				Kind: pkgport.SpanKindInternal,
			})
			if _, err := w.useCases.PublishPendingOutboxEventsUseCase().Execute(tickCtx, struct{}{}); err != nil {
				span.Fail(err)
				w.log.Error("publish pending outbox events failed", "error", err)
			}
			span.End()
		}
	}
}
