package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopLogger(t *testing.T) {
	log := NewNopLogger()
	ctx := context.Background()
	log.Debug(ctx, "d")
	log.Info(ctx, "i")
	log.Warn(ctx, "w")
	log.Error(ctx, "e")
	assert.NoError(t, log.Sync())
}
