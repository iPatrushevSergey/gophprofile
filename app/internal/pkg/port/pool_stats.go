package port

import "context"

// PoolStats provides PostgreSQL connection pool statistics.
type PoolStats interface {
	Stats(ctx context.Context) (DBPoolStats, error)
}
