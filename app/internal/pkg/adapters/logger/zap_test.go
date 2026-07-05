package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZapLogger(t *testing.T) {
	log, err := NewZapLogger(Config{Level: "info", Backend: "zap"})
	require.NoError(t, err)
	log.Info(context.Background(), "test")
	assert.NoError(t, log.Sync())
}

func TestNewZapLogger_invalidLevel(t *testing.T) {
	_, err := NewZapLogger(Config{Level: "not-a-level", Backend: "zap"})
	assert.Error(t, err)
}

func TestZapLogger_keyValuePairs(t *testing.T) {
	log, err := NewZapLogger(Config{Level: "debug", Backend: "zap"})
	require.NoError(t, err)

	ctx := context.Background()
	log.Debug(ctx, "d", "k", "v", 42, "n")
	log.Warn(ctx, "w", "err", errors.New("test err"))
	log.Error(ctx, "e", "n", 1)
	assert.NoError(t, log.Sync())
}
