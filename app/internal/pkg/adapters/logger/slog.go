package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

// SlogLogger implements application logging using log/slog.
type SlogLogger struct {
	sl         *slog.Logger
	otelExport bool
}

// NewSlogLogger builds a slog logger from config.
// When loggerProvider is non-nil, logs are exported via OTLP using serviceName.
func NewSlogLogger(cfg Config, loggerProvider *sdklog.LoggerProvider, serviceName string) (*SlogLogger, error) {
	lvl, err := parseSlogLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	var handler slog.Handler
	switch {
	case loggerProvider != nil:
		handler = otelslog.NewHandler(
			serviceName,
			otelslog.WithLoggerProvider(loggerProvider),
		)
		handler = &minimumLevelHandler{Handler: handler, level: lvl}
	case cfg.Format == "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	case cfg.Format == "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	default:
		return nil, fmt.Errorf("format: unknown value %q", cfg.Format)
	}

	return &SlogLogger{
		sl:         slog.New(handler),
		otelExport: loggerProvider != nil,
	}, nil
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

// appendTraceFields adds trace_id/span_id on stdout; skipped for OTLP (SDK injects from ctx).
func (s *SlogLogger) appendTraceFields(ctx context.Context, args []any) []any {
	if s.otelExport || ctx == nil {
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

// minimumLevelHandler filters logs by level.
type minimumLevelHandler struct {
	slog.Handler
	level slog.Leveler
}

// Enabled checks if the log level is enabled.
func (h *minimumLevelHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// WithAttrs adds attributes to the handler.
func (h *minimumLevelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &minimumLevelHandler{Handler: h.Handler.WithAttrs(attrs), level: h.level}
}

// WithGroup adds a group to the handler.
func (h *minimumLevelHandler) WithGroup(name string) slog.Handler {
	return &minimumLevelHandler{Handler: h.Handler.WithGroup(name), level: h.level}
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
