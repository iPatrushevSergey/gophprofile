package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZapLogger(t *testing.T) {
	log, err := NewZapLogger(Config{Level: "info"})
	require.NoError(t, err)
	log.Info("test")
	assert.NoError(t, log.Sync())
}

func TestNewZapLogger_invalidLevel(t *testing.T) {
	_, err := NewZapLogger(Config{Level: "not-a-level"})
	assert.Error(t, err)
}

func TestZapLogger_keyValuePairs(t *testing.T) {
	log, err := NewZapLogger(Config{Level: "debug"})
	require.NoError(t, err)

	log.Debug("d", "k", "v", 42, "n")
	log.Warn("w", "err", errors.New("test err"))
	log.Error("e", "n", 1)
	assert.NoError(t, log.Sync())
}
