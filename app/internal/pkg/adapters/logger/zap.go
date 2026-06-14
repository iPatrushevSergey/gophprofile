package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// ZapLogger implements application logging using zap.
type ZapLogger struct {
	zl *zap.Logger
}

// NewZapLogger builds a zap.Logger from config.
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

func (z *ZapLogger) Debug(msg string, args ...any) {
	z.zl.Debug(msg, toZapFields(args)...)
}

func (z *ZapLogger) Info(msg string, args ...any) {
	z.zl.Info(msg, toZapFields(args)...)
}

func (z *ZapLogger) Warn(msg string, args ...any) {
	z.zl.Warn(msg, toZapFields(args)...)
}

func (z *ZapLogger) Error(msg string, args ...any) {
	z.zl.Error(msg, toZapFields(args)...)
}

// Sync flushes buffered logs.
func (z *ZapLogger) Sync() error {
	return z.zl.Sync()
}

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
