// Package worker runs background tasks.
package worker

import (
	"context"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// PeriodicMetricsCollectorWorker runs periodic polled metrics collection.
type PeriodicMetricsCollectorWorker struct {
	collectPeriodicMetrics appport.UseCase[struct{}, struct{}]
	log                    pkgport.Logger
	interval               time.Duration
}

// NewPeriodicMetricsCollectorWorker creates a background task to collect polled metrics.
func NewPeriodicMetricsCollectorWorker(
	collectPeriodicMetrics appport.UseCase[struct{}, struct{}],
	log pkgport.Logger,
	interval time.Duration,
) *PeriodicMetricsCollectorWorker {
	return &PeriodicMetricsCollectorWorker{
		collectPeriodicMetrics: collectPeriodicMetrics,
		log:                    log,
		interval:               interval,
	}
}

// Run executes the worker loop.
func (w *PeriodicMetricsCollectorWorker) Run(ctx context.Context) {
	w.log.Info(ctx, "periodic metrics collector worker started", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info(ctx, "periodic metrics collector worker stopped")
			return
		case <-ticker.C:
			workCtx := context.WithoutCancel(ctx)
			if _, err := w.collectPeriodicMetrics.Execute(workCtx, struct{}{}); err != nil {
				w.log.Warn(workCtx, "collect periodic metrics failed", "error", err)
			}
		}
	}
}
