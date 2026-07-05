package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	prommetrics "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/prometheus"
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
	Metrics                        *prommetrics.Metrics
	MetricsAddress                 string
	MetricsEnabled                 bool
	PeriodicMetricsCollectInterval time.Duration
	cancelAvatarProcessorWorker    context.CancelFunc
	cancelPeriodicMetricsCollector context.CancelFunc
	metricsServer                  *http.Server
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

	if a.MetricsEnabled && a.Metrics != nil {
		metricsSrv := &http.Server{
			Addr:    a.MetricsAddress,
			Handler: a.Metrics.Handler(),
		}
		a.metricsServer = metricsSrv
		go func() {
			a.Log.Info(ctx, "metrics server listening", "address", metricsSrv.Addr)
			if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				a.Log.Error(ctx, "metrics server failed", "error", err)
			}
		}()

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

// Stop stops the application.
func (a *App) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.Log.Info(context.Background(), "shutdown signal received, stopping processor...")
	ctx, cancel := context.WithTimeout(context.Background(), a.ShutdownTimeout)
	defer cancel()

	a.cancelAvatarProcessorWorker()

	if a.metricsServer != nil {
		if err := a.metricsServer.Shutdown(ctx); err != nil {
			a.Log.Error(ctx, "metrics server shutdown failed", "error", err)
		}
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

	if a.TelemetryShutdown != nil {
		if err := a.TelemetryShutdown(ctx); err != nil {
			a.Log.Error(ctx, "telemetry shutdown failed", "error", err)
		}
	}

	a.Log.Info(context.Background(), "processor stopped gracefully")
	return nil
}
