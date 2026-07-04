package port

import "context"

// DBPoolStatsReader provides PostgreSQL connection pool statistics.
type DBPoolStatsReader interface {
	Stats(ctx context.Context) (DBPoolStats, error)
}
