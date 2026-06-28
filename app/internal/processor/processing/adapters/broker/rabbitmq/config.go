package rabbitmq

import (
	"fmt"
	"strings"
	"time"
)

// Config holds RabbitMQ consumer settings.
type Config struct {
	URL                   string        `mapstructure:"url"`
	Exchange              string        `mapstructure:"exchange"`
	Queue                 string        `mapstructure:"queue"`
	DeadLetterExchange    string        `mapstructure:"dead_letter_exchange"`
	DeadLetterQueue       string        `mapstructure:"dead_letter_queue"`
	DeadLetterRoutingKey  string        `mapstructure:"dead_letter_routing_key"`
	RetryQueue            string        `mapstructure:"retry_queue"`
	RetryExchange         string        `mapstructure:"retry_exchange"`
	RetryReturnRoutingKey string        `mapstructure:"retry_return_routing_key"`
	RetryTTL              time.Duration `mapstructure:"retry_ttl"`
	MaxRetries            int           `mapstructure:"max_retries"`
	ReconnectInterval     time.Duration `mapstructure:"reconnect_interval"`
}

// Validate trims string fields and checks broker settings when a URL is configured.
func (c *Config) Validate() error {
	c.URL = strings.TrimSpace(c.URL)
	c.Exchange = strings.TrimSpace(c.Exchange)
	c.Queue = strings.TrimSpace(c.Queue)
	c.DeadLetterExchange = strings.TrimSpace(c.DeadLetterExchange)
	c.DeadLetterQueue = strings.TrimSpace(c.DeadLetterQueue)
	c.DeadLetterRoutingKey = strings.TrimSpace(c.DeadLetterRoutingKey)
	c.RetryQueue = strings.TrimSpace(c.RetryQueue)
	c.RetryExchange = strings.TrimSpace(c.RetryExchange)
	c.RetryReturnRoutingKey = strings.TrimSpace(c.RetryReturnRoutingKey)

	if !c.Enabled() {
		return nil
	}

	if c.Exchange == "" {
		return fmt.Errorf("broker exchange is required")
	}

	if c.Queue == "" {
		return fmt.Errorf("broker queue is required")
	}

	if c.RetryQueue == "" {
		return fmt.Errorf("broker retry_queue is required")
	}

	if c.RetryExchange == "" {
		return fmt.Errorf("broker retry_exchange is required")
	}

	if c.RetryReturnRoutingKey == "" {
		return fmt.Errorf("broker retry_return_routing_key is required")
	}

	if c.RetryTTL < time.Millisecond {
		return fmt.Errorf("broker retry_ttl must be at least 1ms")
	}

	if c.MaxRetries <= 0 {
		return fmt.Errorf("broker max_retries must be positive")
	}

	if c.ReconnectInterval <= 0 {
		return fmt.Errorf("broker reconnect_interval must be positive")
	}

	return nil
}

// Enabled reports whether RabbitMQ integration is configured.
func (c Config) Enabled() bool {
	return c.URL != ""
}
