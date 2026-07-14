package main

import (
	"log"
	"os"

	"github.com/iPatrushevSergey/gophprofile/app/cmd/server/bootstrap"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
)

func main() {
	if apputil.HandleVersionArg(os.Args[1:]) {
		return
	}

	if err := bootstrap.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
