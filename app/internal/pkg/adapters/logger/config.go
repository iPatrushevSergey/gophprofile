package logger

import (
	"fmt"
	"strings"
)

// Config defines logger initialization parameters.
type Config struct {
	Level string `mapstructure:"level"`
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
	return nil
}

func isKnownLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error", "dpanic", "panic", "fatal":
		return true
	default:
		return false
	}
}
