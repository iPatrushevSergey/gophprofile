//go:build integration || e2e || component || contract

package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBName = "gophprofile_test"
	testDBUser = "test"
	testDBPass = "test"
)

// SetupPostgres starts PostgreSQL, applies migrations and returns a pgxpool.Pool.
func SetupPostgres(tb testing.TB) *pgxpool.Pool {
	tb.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(tb, err, "start postgres container")

	tb.Cleanup(func() {
		require.NoError(tb, pgContainer.Terminate(ctx))
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(tb, err)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(tb, err)
	poolCfg.MaxConns = 5
	poolCfg.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	require.NoError(tb, err)
	tb.Cleanup(func() { pool.Close() })

	require.NoError(tb, migrate.PostgresUp(dsn, migrate.MigrationsGophprofileDir()))

	return pool
}
