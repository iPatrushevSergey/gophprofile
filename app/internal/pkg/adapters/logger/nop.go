package logger

import "context"

// NopLogger is a no-op logger for tests.
type NopLogger struct{}

// NewNopLogger returns a logger that discards all output.
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// Debug implements the Debug method of the Logger interface.
func (NopLogger) Debug(ctx context.Context, msg string, args ...any) {}

// Info implements the Info method of the Logger interface.
func (NopLogger) Info(ctx context.Context, msg string, args ...any) {}

// Warn implements the Warn method of the Logger interface.
func (NopLogger) Warn(ctx context.Context, msg string, args ...any) {}

// Error implements the Error method of the Logger interface.
func (NopLogger) Error(ctx context.Context, msg string, args ...any) {}

// Sync implements the Sync method of the Logger interface.
func (NopLogger) Sync() error { return nil }
