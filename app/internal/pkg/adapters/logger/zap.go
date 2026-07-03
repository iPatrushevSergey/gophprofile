package logger

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ZapLogger implements application logging using zap.
type ZapLogger struct {
	zl *zap.Logger
}

// NewZapLogger builds a zap logger from config.
func NewZapLogger(cfg Config) (*ZapLogger, error) {
	lvl, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = lvl
	zl, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{zl: zl}, nil
}

// Debug implements the Debug method of the Logger interface.
func (z *ZapLogger) Debug(ctx context.Context, msg string, args ...any) {
	z.zl.Debug(msg, toZapFields(z.appendTraceFields(ctx, args))...)
}

// Info implements the Info method of the Logger interface.
func (z *ZapLogger) Info(ctx context.Context, msg string, args ...any) {
	z.zl.Info(msg, toZapFields(z.appendTraceFields(ctx, args))...)
}

// Warn implements the Warn method of the Logger interface.
func (z *ZapLogger) Warn(ctx context.Context, msg string, args ...any) {
	z.zl.Warn(msg, toZapFields(z.appendTraceFields(ctx, args))...)
}

// Error implements the Error method of the Logger interface.
func (z *ZapLogger) Error(ctx context.Context, msg string, args ...any) {
	z.zl.Error(msg, toZapFields(z.appendTraceFields(ctx, args))...)
}

// Sync flushes buffered logs.
func (z *ZapLogger) Sync() error {
	return z.zl.Sync()
}

// appendTraceFields appends trace fields to the arguments.
func (z *ZapLogger) appendTraceFields(ctx context.Context, args []any) []any {
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

// toZapFields converts a slice of any to a slice of zap.Field.
func toZapFields(args []any) []zap.Field {
	if len(args) == 0 {
		return nil
	}
	fields := make([]zap.Field, 0, len(args)/2+1)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("key%d", i/2)
		}
		fields = append(fields, toZapField(key, args[i+1]))
	}
	return fields
}

// toZapField converts a key and value to a zap.Field.
func toZapField(key string, val any) zap.Field {
	switch v := val.(type) {
	case error:
		return zap.Error(v)
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case bool:
		return zap.Bool(key, v)
	default:
		return zap.Any(key, val)
	}
}
