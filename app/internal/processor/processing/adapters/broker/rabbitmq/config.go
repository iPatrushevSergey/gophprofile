package rabbitmq

import (
	"fmt"
	"strings"
)

// Config holds RabbitMQ consumer settings.
type Config struct {
	URL                  string `mapstructure:"url"`
	Exchange             string `mapstructure:"exchange"`
	Queue                string `mapstructure:"queue"`
	DeadLetterExchange   string `mapstructure:"dead_letter_exchange"`
	DeadLetterQueue      string `mapstructure:"dead_letter_queue"`
	DeadLetterRoutingKey string `mapstructure:"dead_letter_routing_key"`
}

// Validate trims string fields and checks broker settings when a URL is configured.
func (c *Config) Validate() error {
	c.URL = strings.TrimSpace(c.URL)
	c.Exchange = strings.TrimSpace(c.Exchange)
	c.Queue = strings.TrimSpace(c.Queue)
	c.DeadLetterExchange = strings.TrimSpace(c.DeadLetterExchange)
	c.DeadLetterQueue = strings.TrimSpace(c.DeadLetterQueue)
	c.DeadLetterRoutingKey = strings.TrimSpace(c.DeadLetterRoutingKey)

	if !c.Enabled() {
		return nil
	}

	if c.Exchange == "" {
		return fmt.Errorf("broker exchange is required")
	}

	if c.Queue == "" {
		return fmt.Errorf("broker queue is required")
	}

	return nil
}

// Enabled reports whether RabbitMQ integration is configured.
func (c Config) Enabled() bool {
	return c.URL != ""
}
