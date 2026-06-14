package logger

import "strings"

// Config defines logger initialization parameters.
type Config struct {
	Level string `mapstructure:"level"`
}

// Normalize trims string fields after config load.
func (c *Config) Normalize() {
	c.Level = strings.TrimSpace(c.Level)
}
