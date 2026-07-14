package logger

import (
	"fmt"
	"strings"
)

// Config defines logger initialization parameters.
type Config struct {
	Level   string `mapstructure:"level"`
	Backend string `mapstructure:"backend"`
	Format  string `mapstructure:"format"`
}

// Validate trims string fields and checks logger settings.
func (c *Config) Validate() error {
	c.Level = strings.TrimSpace(c.Level)
	if c.Level == "" {
		return fmt.Errorf("level is required")
	}
	if !isKnownLevel(c.Level) {
		return fmt.Errorf("level: unknown value %q", c.Level)
	}

	c.Backend = strings.ToLower(strings.TrimSpace(c.Backend))
	if c.Backend == "" {
		return fmt.Errorf("backend is required")
	}
	switch c.Backend {
	case "zap", "slog":
	default:
		return fmt.Errorf("backend: unknown value %q", c.Backend)
	}

	c.Format = strings.ToLower(strings.TrimSpace(c.Format))
	if c.Backend == "slog" {
		if c.Format == "" {
			return fmt.Errorf("format is required")
		}
		switch c.Format {
		case "json", "text":
		default:
			return fmt.Errorf("format: unknown value %q", c.Format)
		}
	}

	return nil
}

// isKnownLevel checks if the level is known.
func isKnownLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error", "dpanic", "panic", "fatal":
		return true
	default:
		return false
	}
}
