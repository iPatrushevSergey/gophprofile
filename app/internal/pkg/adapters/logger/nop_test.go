package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopLogger(t *testing.T) {
	log := NewNopLogger()
	log.Debug("d")
	log.Info("i")
	log.Warn("w")
	log.Error("e")
	assert.NoError(t, log.Sync())
}
