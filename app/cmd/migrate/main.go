package main

import (
	"flag"
	"log"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/migrate"
)

func main() {
	dsn := flag.String("d", "", "database DSN")
	dir := flag.String("dir", migrate.MigrationsGophprofileDir(), "path to migration files")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("migrate: -d is required")
	}

	if err := migrate.PostgresUp(*dsn, *dir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
