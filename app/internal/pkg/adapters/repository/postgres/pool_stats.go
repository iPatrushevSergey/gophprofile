package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// PoolStats reads connection pool statistics from pgxpool.
type PoolStats struct {
	pool *pgxpool.Pool
}

var _ pkgport.PoolStats = (*PoolStats)(nil)

// NewPoolStats creates a pool statistics provider.
func NewPoolStats(pool *pgxpool.Pool) *PoolStats {
	return &PoolStats{pool: pool}
}

// Stats returns current pool connection statistics.
func (s *PoolStats) Stats(_ context.Context) (pkgport.DBPoolStats, error) {
	if s.pool == nil {
		return pkgport.DBPoolStats{}, nil
	}

	stat := s.pool.Stat()
	return pkgport.DBPoolStats{
		TotalConns:    stat.TotalConns(),
		IdleConns:     stat.IdleConns(),
		AcquiredConns: stat.AcquiredConns(),
	}, nil
}
