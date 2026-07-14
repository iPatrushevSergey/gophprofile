package otel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("disabled skips collect interval validation", func(t *testing.T) {
		cfg := Config{Enabled: false}
		require.NoError(t, cfg.Validate())
	})

	t.Run("enabled requires positive collect interval", func(t *testing.T) {
		cfg := Config{Enabled: true}
		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "collect_interval must be positive")
	})

	t.Run("enabled with collect interval is valid", func(t *testing.T) {
		cfg := Config{Enabled: true, CollectInterval: 15 * time.Second}
		require.NoError(t, cfg.Validate())
	})
}
