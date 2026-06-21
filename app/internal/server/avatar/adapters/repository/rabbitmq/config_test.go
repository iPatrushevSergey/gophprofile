package rabbitmq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	cfg := Config{
		URL:             " amqp://localhost:5672 ",
		Exchange:        " avatars ",
		PublishInterval: 5 * time.Second,
		OutboxBatchSize: 100,
	}

	require.NoError(t, cfg.Validate())
	assert.Equal(t, "amqp://localhost:5672", cfg.URL)
	assert.Equal(t, "avatars", cfg.Exchange)
}

func TestConfig_Validate_skipsWhenDisabled(t *testing.T) {
	cfg := Config{}

	require.NoError(t, cfg.Validate())
}

func TestConfig_Validate_requiresExchangeWhenEnabled(t *testing.T) {
	cfg := Config{
		URL:             "amqp://localhost:5672",
		PublishInterval: 5 * time.Second,
		OutboxBatchSize: 100,
	}

	require.Error(t, cfg.Validate())
}
