package otel

import (
	"fmt"
	"time"
)

// Config holds OpenTelemetry metrics settings.
type Config struct {
	Enabled         bool          `mapstructure:"enabled"`
	CollectInterval time.Duration `mapstructure:"collect_interval"`
}

// Validate checks metrics settings.
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.CollectInterval <= 0 {
		return fmt.Errorf("collect_interval must be positive")
	}
	return nil
}
