package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewSlogLogger(t *testing.T) {
	log, err := NewSlogLogger(Config{Level: "info", Backend: "slog", Format: "json"}, nil, "")
	require.NoError(t, err)
	log.Info(context.Background(), "test")
	assert.NoError(t, log.Sync())
}

func TestNewSlogLogger_textFormat(t *testing.T) {
	log, err := NewSlogLogger(Config{Level: "debug", Backend: "slog", Format: "text"}, nil, "")
	require.NoError(t, err)
	log.Debug(context.Background(), "test")
	assert.NoError(t, log.Sync())
}

func TestNewSlogLogger_invalidLevel(t *testing.T) {
	_, err := NewSlogLogger(Config{Level: "not-a-level", Backend: "slog", Format: "json"}, nil, "")
	assert.Error(t, err)
}

func TestNewSlogLogger_invalidFormat(t *testing.T) {
	_, err := NewSlogLogger(Config{Level: "info", Backend: "slog", Format: "yaml"}, nil, "")
	assert.Error(t, err)
}

func TestSlogLogger_keyValuePairs(t *testing.T) {
	log, err := NewSlogLogger(Config{Level: "debug", Backend: "slog", Format: "json"}, nil, "")
	require.NoError(t, err)

	log.Debug(context.Background(), "d", "k", "v", 42, "n")
	log.Warn(context.Background(), "w", "err", errors.New("test err"))
	log.Error(context.Background(), "e", "n", 1)
	assert.NoError(t, log.Sync())
}

func TestConfigValidate_backendAndFormat(t *testing.T) {
	cfg := Config{Level: "info", Backend: "slog", Format: "text"}
	require.NoError(t, cfg.Validate())
	assert.Equal(t, "slog", cfg.Backend)
	assert.Equal(t, "text", cfg.Format)

	cfg = Config{Level: "info", Backend: "slog", Format: "xml"}
	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_missingBackend(t *testing.T) {
	cfg := Config{Level: "info"}
	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_missingFormat(t *testing.T) {
	cfg := Config{Level: "info", Backend: "slog"}
	assert.Error(t, cfg.Validate())
}

func TestConfigValidate_invalidBackend(t *testing.T) {
	cfg := Config{Level: "info", Backend: "logrus"}
	assert.Error(t, cfg.Validate())
}

func TestSlogLogger_appendTraceFields_withoutSpan(t *testing.T) {
	log := &SlogLogger{otelExport: false}
	args := log.appendTraceFields(context.Background(), []any{"error", assert.AnError})
	assert.Equal(t, []any{"error", assert.AnError}, args)
}

func TestSlogLogger_appendTraceFields_withSpan(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tr := tp.Tracer("test")
	ctx, span := tr.Start(context.Background(), "test")
	defer span.End()

	log := &SlogLogger{otelExport: false}
	args := log.appendTraceFields(ctx, []any{"error", assert.AnError})
	require.GreaterOrEqual(t, len(args), 4)
	assert.Equal(t, "trace_id", args[0])
	assert.NotEmpty(t, args[1])
	assert.Equal(t, "span_id", args[2])
	assert.NotEmpty(t, args[3])
	assert.Equal(t, "error", args[4])
}
