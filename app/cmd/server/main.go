package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iPatrushevSergey/gophprofile/app/cmd/server/bootstrap"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
)

func main() {
	if apputil.HandleVersionArg(os.Args[1:]) {
		return
	}

	app, cleanups, err := bootstrap.Run()
	for _, cleanup := range cleanups {
		defer cleanup()
	}
	if err != nil {
		log.Fatalf("server: %v", err)
	}
	if app == nil {
		return
	}

	app.Start()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	<-signalCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), app.ShutdownTimeout)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}
