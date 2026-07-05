package prometheus

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Config holds Prometheus metrics settings.
type Config struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

// Validate trims and checks metrics settings.
func (c *Config) Validate() error {
	c.Address = strings.TrimSpace(c.Address)
	if !c.Enabled {
		return nil
	}
	if c.Address == "" {
		return fmt.Errorf("address is required when metrics are enabled")
	}
	_, portStr, err := net.SplitHostPort(c.Address)
	if err != nil {
		return fmt.Errorf("address: %w", err)
	}
	if _, err := strconv.Atoi(portStr); err != nil {
		return fmt.Errorf("address: invalid port: %w", err)
	}
	return nil
}
