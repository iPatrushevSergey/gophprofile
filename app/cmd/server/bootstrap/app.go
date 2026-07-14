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

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	avatarworker "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/worker"
)

// App represents the application lifecycle.
type App struct {
	Server            *http.Server
	Log               pkgport.Logger
	TelemetryShutdown func(context.Context) error
	ShutdownTimeout   time.Duration
	TLSCertFile       string
	TLSKeyFile        string

	UseCases                       GlobalUseCases
	Tracer                         pkgport.Tracer
	MetricsEnabled                 bool
	PeriodicMetricsCollectInterval time.Duration
	UploadGCInterval               time.Duration
	OutboxPublishInterval          time.Duration
	cancelUploadGCWorker           context.CancelFunc
	cancelOutboxPublisherWorker    context.CancelFunc
	cancelPeriodicMetricsCollector context.CancelFunc
	workerWg                       sync.WaitGroup
}

// Start starts the application.
func (a *App) Start() {
	ctx := context.Background()

	if a.UploadGCInterval > 0 {
		workerCtx, cancel := context.WithCancel(ctx)
		a.cancelUploadGCWorker = cancel
		a.workerWg.Add(1)
		go func() {
			defer a.workerWg.Done()
			avatarworker.NewUploadingAvatarGCWorker(
				a.UseCases,
				a.Log,
				a.UploadGCInterval,
			).Run(workerCtx)
		}()
	}

	if a.OutboxPublishInterval > 0 {
		workerCtx, cancel := context.WithCancel(ctx)
		a.cancelOutboxPublisherWorker = cancel
		a.workerWg.Add(1)
		go func() {
			defer a.workerWg.Done()
			avatarworker.NewOutboxPublisherWorker(
				a.UseCases,
				a.Log,
				a.Tracer,
				a.OutboxPublishInterval,
			).Run(workerCtx)
		}()
	}

	if a.MetricsEnabled {
		metricsCtx, cancel := context.WithCancel(ctx)
		a.cancelPeriodicMetricsCollector = cancel
		a.workerWg.Add(1)
		go func() {
			defer a.workerWg.Done()
			avatarworker.NewPeriodicMetricsCollectorWorker(
				a.UseCases.CollectPeriodicMetricsUseCase(),
				a.Log,
				a.PeriodicMetricsCollectInterval,
			).Run(metricsCtx)
		}()
	}

	go func() {
		a.Log.Info(ctx, "server listening", "address", a.Server.Addr, "tls", a.TLSCertFile != "")
		var err error
		if a.TLSCertFile != "" && a.TLSKeyFile != "" {
			err = a.Server.ListenAndServeTLS(a.TLSCertFile, a.TLSKeyFile)
		} else {
			err = a.Server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Log.Error(context.Background(), "server failed", "error", err)
		}
	}()
}

// Stop stops the application.
func (a *App) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.Log.Info(context.Background(), "shutdown signal received, stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), a.ShutdownTimeout)
	defer cancel()

	var shutdownErr error
	if err := a.Server.Shutdown(ctx); err != nil {
		a.Log.Error(ctx, "server shutdown failed", "error", err)
		shutdownErr = err
	}

	if a.cancelUploadGCWorker != nil {
		a.cancelUploadGCWorker()
	}
	if a.cancelOutboxPublisherWorker != nil {
		a.cancelOutboxPublisherWorker()
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

	if shutdownErr == nil {
		a.Log.Info(context.Background(), "server stopped gracefully")
	}

	if a.TelemetryShutdown != nil {
		if err := a.TelemetryShutdown(ctx); err != nil {
			a.Log.Error(ctx, "telemetry shutdown failed", "error", err)
		}
	}

	return shutdownErr
}
