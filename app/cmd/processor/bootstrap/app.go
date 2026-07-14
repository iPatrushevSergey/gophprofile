package bootstrap

import (
	"context"
	"errors"
	"sync"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	processingappport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	processingworker "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/presentation/worker"
)

// App represents the application lifecycle.
type App struct {
	Log               pkgport.Logger
	TelemetryShutdown func(context.Context) error
	ShutdownTimeout   time.Duration

	UseCases                       GlobalUseCases
	EventConsumer                  processingappport.EventConsumer
	MetricsEnabled                 bool
	PeriodicMetricsCollectInterval time.Duration
	HealthFileInterval             time.Duration
	cancelAvatarProcessorWorker    context.CancelFunc
	cancelPeriodicMetricsCollector context.CancelFunc
	cancelHealthFileWorker         context.CancelFunc
	workerWg                       sync.WaitGroup
}

// Start starts the application.
func (a *App) Start() {
	ctx := context.Background()

	workerCtx, cancel := context.WithCancel(ctx)
	a.cancelAvatarProcessorWorker = cancel
	a.workerWg.Add(1)
	go func() {
		defer a.workerWg.Done()

		if err := processingworker.NewAvatarProcessorWorker(
			a.UseCases,
			a.Log,
		).Run(workerCtx); err != nil && !errors.Is(err, context.Canceled) {
			a.Log.Error(workerCtx, "processor worker failed", "error", err)
		}
	}()

	healthCtx, healthCancel := context.WithCancel(ctx)
	a.cancelHealthFileWorker = healthCancel
	a.workerWg.Add(1)
	go func() {
		defer a.workerWg.Done()
		processingworker.NewHealthFileWorker(
			a.UseCases.RefreshHealthFileUseCase(),
			a.Log,
			a.HealthFileInterval,
		).Run(healthCtx)
	}()

	if a.MetricsEnabled {
		metricsCtx, cancel := context.WithCancel(ctx)
		a.cancelPeriodicMetricsCollector = cancel
		a.workerWg.Add(1)
		go func() {
			defer a.workerWg.Done()
			processingworker.NewPeriodicMetricsCollectorWorker(
				a.UseCases.CollectPeriodicMetricsUseCase(),
				a.Log,
				a.PeriodicMetricsCollectInterval,
			).Run(metricsCtx)
		}()
	}

	a.Log.Info(ctx, "processor worker started")
}

// Shutdown stops background workers, the event consumer and telemetry.
func (a *App) Shutdown(ctx context.Context) error {
	a.Log.Info(context.Background(), "stopping processor...")

	a.cancelAvatarProcessorWorker()

	if a.cancelHealthFileWorker != nil {
		a.cancelHealthFileWorker()
	}

	if a.cancelPeriodicMetricsCollector != nil {
		a.cancelPeriodicMetricsCollector()
	}

	workersDone := make(chan struct{})
	go func() {
		a.workerWg.Wait()
		close(workersDone)
	}()
	select {
	case <-workersDone:
	case <-ctx.Done():
		a.Log.Warn(ctx, "background workers shutdown timeout exceeded", "timeout", a.ShutdownTimeout)
	}

	if a.EventConsumer != nil {
		if err := a.EventConsumer.Close(); err != nil {
			a.Log.Error(ctx, "close event consumer failed", "error", err)
		}
	}

	a.Log.Info(context.Background(), "processor stopped gracefully")

	if a.TelemetryShutdown != nil {
		if err := a.TelemetryShutdown(ctx); err != nil {
			a.Log.Error(ctx, "telemetry shutdown failed", "error", err)
		}
	}

	return nil
}
