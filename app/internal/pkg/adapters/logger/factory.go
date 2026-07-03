package logger

import (
	"fmt"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// NewLogger builds a logger from config using the selected backend.
func NewLogger(cfg Config) (pkgport.Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	switch cfg.Backend {
	case "zap":
		return NewZapLogger(cfg)
	case "slog":
		return NewSlogLogger(cfg)
	default:
		return nil, fmt.Errorf("backend: unknown value %q", cfg.Backend)
	}
}
