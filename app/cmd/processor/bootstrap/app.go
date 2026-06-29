package bootstrap

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
	processingworker "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/presentation/worker"
)

// App represents the application lifecycle.
type App struct {
	Log             pkgport.Logger
	ShutdownTimeout time.Duration

	UseCases                    GlobalUseCases
	EventConsumer               appport.EventConsumer
	cancelAvatarProcessorWorker context.CancelFunc
	workerDone                  chan struct{}
}

// Start starts the application.
func (a *App) Start() {
	ctx := context.Background()

	workerCtx, cancel := context.WithCancel(ctx)
	a.cancelAvatarProcessorWorker = cancel
	a.workerDone = make(chan struct{})

	go func() {
		defer close(a.workerDone)

		if err := processingworker.NewAvatarProcessorWorker(
			a.UseCases,
			a.Log,
		).Run(workerCtx); err != nil && !errors.Is(err, context.Canceled) {
			a.Log.Error("processor worker failed", "error", err)
		}
	}()

	a.Log.Info("processor worker started")
}

// Stop stops the application.
func (a *App) Stop() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	a.Log.Info("shutdown signal received, stopping processor...")
	ctx, cancel := context.WithTimeout(context.Background(), a.ShutdownTimeout)
	defer cancel()

	a.cancelAvatarProcessorWorker()

	select {
	case <-a.workerDone:
	case <-ctx.Done():
		a.Log.Warn("processor worker shutdown timeout exceeded", "timeout", a.ShutdownTimeout)
	}

	if a.EventConsumer != nil {
		if err := a.EventConsumer.Close(); err != nil {
			a.Log.Error("close event consumer failed", "error", err)
		}
	}

	a.Log.Info("processor stopped gracefully")
	return nil
}
