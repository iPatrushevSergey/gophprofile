package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig defines PostgreSQL pool settings.
type PoolConfig struct {
	URI         string        `mapstructure:"uri"`
	MaxConns    int32         `mapstructure:"max_conns"`
	MinConns    int32         `mapstructure:"min_conns"`
	MaxConnLife time.Duration `mapstructure:"max_conn_life"`
	MaxConnIdle time.Duration `mapstructure:"max_conn_idle"`
	HealthCheck time.Duration `mapstructure:"health_check"`
}

// RetryConfig defines transactor retry settings for retriable DB errors.
type RetryConfig struct {
	MaxRetries int           `mapstructure:"max_retries"`
	BaseDelay  time.Duration `mapstructure:"base_delay"`
	MaxDelay   time.Duration `mapstructure:"max_delay"`
}

// Config groups adapter-level postgres settings.
type Config struct {
	Pool  PoolConfig  `mapstructure:",squash"`
	Retry RetryConfig `mapstructure:"retry"`
}

// Validate trims string fields and checks postgres settings when a URI is configured.
func (c *Config) Validate() error {
	c.Pool.URI = strings.TrimSpace(c.Pool.URI)
	if c.Pool.URI == "" {
		return nil
	}

	if c.Pool.MaxConns <= 0 {
		return fmt.Errorf("max_conns must be positive")
	}

	if c.Pool.MinConns < 0 {
		return fmt.Errorf("min_conns must be non-negative")
	}

	if c.Pool.MinConns > c.Pool.MaxConns {
		return fmt.Errorf("min_conns must not exceed max_conns")
	}

	if c.Pool.MaxConnLife <= 0 {
		return fmt.Errorf("max_conn_life must be positive")
	}

	if c.Pool.MaxConnIdle <= 0 {
		return fmt.Errorf("max_conn_idle must be positive")
	}

	if c.Pool.HealthCheck <= 0 {
		return fmt.Errorf("health_check must be positive")
	}

	if c.Retry.MaxRetries < 0 {
		return fmt.Errorf("retry max_retries must be non-negative")
	}

	if c.Retry.BaseDelay <= 0 {
		return fmt.Errorf("retry base_delay must be positive")
	}

	if c.Retry.MaxDelay <= 0 {
		return fmt.Errorf("retry max_delay must be positive")
	}

	if c.Retry.MaxDelay < c.Retry.BaseDelay {
		return fmt.Errorf("retry max_delay must not be less than base_delay")
	}
	return nil
}

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
