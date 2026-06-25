package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	avatarworker "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/worker"
)

// App represents the application lifecycle.
type App struct {
	Server          *http.Server
	Log             appport.Logger
	ShutdownTimeout time.Duration
	TLSCertFile     string
	TLSKeyFile      string

	UseCases                    GlobalUseCases
	UploadGCInterval            time.Duration
	OutboxPublishInterval       time.Duration
	cancelUploadGCWorker        context.CancelFunc
	cancelOutboxPublisherWorker context.CancelFunc
}

// Start starts the application.
func (a *App) Start() {
	ctx := context.Background()

	if a.UploadGCInterval > 0 {
		workerCtx, cancel := context.WithCancel(ctx)
		a.cancelUploadGCWorker = cancel
		go avatarworker.NewUploadingAvatarGCWorker(
			a.UseCases,
			a.Log,
			a.UploadGCInterval,
		).Run(workerCtx)
	}

	if a.OutboxPublishInterval > 0 {
		workerCtx, cancel := context.WithCancel(ctx)
		a.cancelOutboxPublisherWorker = cancel
		go avatarworker.NewOutboxPublisherWorker(
			a.UseCases,
			a.Log,
			a.OutboxPublishInterval,
		).Run(workerCtx)
	}

	go func() {
		a.Log.Info("server listening", "address", a.Server.Addr, "tls", a.TLSCertFile != "")
		var err error
		if a.TLSCertFile != "" && a.TLSKeyFile != "" {
			err = a.Server.ListenAndServeTLS(a.TLSCertFile, a.TLSKeyFile)
		} else {
			err = a.Server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Log.Error("server failed", "error", err)
		}
	}()
}

// Stop stops the application.
func (a *App) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.Log.Info("shutdown signal received, stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), a.ShutdownTimeout)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Log.Error("server shutdown failed", "error", err)
		return err
	}

	if a.cancelUploadGCWorker != nil {
		a.cancelUploadGCWorker()
	}
	if a.cancelOutboxPublisherWorker != nil {
		a.cancelOutboxPublisherWorker()
	}

	a.Log.Info("server stopped gracefully")
	return nil
}
