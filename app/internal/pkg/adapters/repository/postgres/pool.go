package postgres

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// NewPool creates a pgxpool.Pool from postgres adapter config.
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("parse database URI: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLife
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdle
	poolCfg.HealthCheckPeriod = cfg.HealthCheck
	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// PoolStatsReader reads connection pool statistics from pgxpool.
type PoolStatsReader struct {
	pool *pgxpool.Pool
}

var _ pkgport.DBPoolStatsReader = (*PoolStatsReader)(nil)

// NewPoolStatsReader creates a DB pool stats reader.
func NewPoolStatsReader(pool *pgxpool.Pool) *PoolStatsReader {
	return &PoolStatsReader{pool: pool}
}

// Stats returns current pool connection statistics.
func (r *PoolStatsReader) Stats(_ context.Context) (pkgport.DBPoolStats, error) {
	if r.pool == nil {
		return pkgport.DBPoolStats{}, nil
	}

	stat := r.pool.Stat()
	return pkgport.DBPoolStats{
		TotalConns:    stat.TotalConns(),
		IdleConns:     stat.IdleConns(),
		AcquiredConns: stat.AcquiredConns(),
	}, nil
}
