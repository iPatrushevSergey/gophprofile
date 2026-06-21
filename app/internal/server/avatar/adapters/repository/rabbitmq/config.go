package rabbitmq

import (
	"fmt"
	"strings"
	"time"
)

// Config holds RabbitMQ publisher settings.
type Config struct {
	URL             string        `mapstructure:"url"`
	Exchange        string        `mapstructure:"exchange"`
	PublishInterval time.Duration `mapstructure:"publish_interval"`
	OutboxBatchSize int           `mapstructure:"outbox_batch_size"`
}

// Validate trims string fields and checks broker settings when a URL is configured.
func (c *Config) Validate() error {
	c.URL = strings.TrimSpace(c.URL)
	c.Exchange = strings.TrimSpace(c.Exchange)

	if !c.Enabled() {
		return nil
	}

	if c.Exchange == "" {
		return fmt.Errorf("broker exchange is required")
	}

	if c.PublishInterval <= 0 {
		return fmt.Errorf("broker publish_interval must be positive")
	}

	if c.OutboxBatchSize <= 0 {
		return fmt.Errorf("broker outbox_batch_size must be positive")
	}
	return nil
}

// Enabled reports whether RabbitMQ integration is configured.
func (c Config) Enabled() bool {
	return c.URL != ""
}
