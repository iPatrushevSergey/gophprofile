package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// SlogLogger implements application logging using log/slog.
type SlogLogger struct {
	sl *slog.Logger
}

// NewSlogLogger builds a slog logger from config.
func NewSlogLogger(cfg Config) (*SlogLogger, error) {
	lvl, err := parseSlogLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		return nil, fmt.Errorf("format: unknown value %q", cfg.Format)
	}

	return &SlogLogger{sl: slog.New(handler)}, nil
}

// Debug implements the Debug method of the Logger interface.
func (s *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	s.sl.Log(ctx, slog.LevelDebug, msg, s.appendTraceFields(ctx, args)...)
}

// Info implements the Info method of the Logger interface.
func (s *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	s.sl.Log(ctx, slog.LevelInfo, msg, s.appendTraceFields(ctx, args)...)
}

// Warn implements the Warn method of the Logger interface.
func (s *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	s.sl.Log(ctx, slog.LevelWarn, msg, s.appendTraceFields(ctx, args)...)
}

// Error implements the Error method of the Logger interface.
func (s *SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	s.sl.Log(ctx, slog.LevelError, msg, s.appendTraceFields(ctx, args)...)
}

// Sync flushes buffered logs.
func (s *SlogLogger) Sync() error {
	return nil
}

// appendTraceFields appends trace fields to the arguments.
func (s *SlogLogger) appendTraceFields(ctx context.Context, args []any) []any {
	if ctx == nil {
		return args
	}

	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if !spanContext.IsValid() {
		return args
	}

	return append([]any{
		"trace_id", spanContext.TraceID().String(),
		"span_id", spanContext.SpanID().String(),
	}, args...)
}

// parseSlogLevel parses a slog level from a string.
func parseSlogLevel(level string) (slog.Leveler, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error", "dpanic", "panic", "fatal":
		return slog.LevelError, nil
	default:
		return nil, fmt.Errorf("parse log level: unknown level %q", level)
	}
}
