package port

import "context"

// Logger provides structured logging. Args are key-value pairs (e.g. "error", err).
type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Sync() error
}
