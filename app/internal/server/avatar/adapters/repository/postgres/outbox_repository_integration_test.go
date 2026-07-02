//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	postgreskit "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/retry"
	oteltelemetry "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/testutil"
	avatarpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
)

func setupOutboxRepo(tb testing.TB) *avatarpostgres.OutboxRepository {
	tb.Helper()
	pool := testutil.SetupPostgres(tb)
	tx := postgreskit.NewTransactor(pool, retry.WithMaxRetries(0))
	tb.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `TRUNCATE avatar_outbox RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	_, err := pool.Exec(context.Background(), `TRUNCATE avatar_outbox RESTART IDENTITY CASCADE`)
	require.NoError(tb, err)
	return avatarpostgres.NewOutboxRepository(tx)
}

func TestOutboxRepository_CreateUploadedAndMarkPublishing(t *testing.T) {
	repo := setupOutboxRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	outboxID := uuid.NewString()
	require.NoError(t, repo.CreateUploaded(ctx, dto.OutboxUploadedCreate{
		ID:        outboxID,
		CreatedAt: now,
		Event: dto.AvatarUploadedEvent{
			AvatarID: "avatar-1",
			UserID:   "user-1",
			S3Key:    "user-1/avatar-1/original",
		},
	}))

	events, err := repo.MarkPublishing(ctx, 10, now)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, outboxID, events[0].ID)
	assert.Equal(t, "avatar-1", events[0].AvatarID)
}

func TestOutboxRepository_persistsTraceCarrier(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	repo := setupOutboxRepo(t)
	ctx, span := otel.Tracer("test").Start(context.Background(), "upload")
	span.End()

	outboxID := uuid.NewString()
	now := time.Now().UTC().Truncate(time.Microsecond)
	carrier := oteltelemetry.NewTracer().ContextToMap(ctx)

	require.NoError(t, repo.CreateUploaded(ctx, dto.OutboxUploadedCreate{
		ID:           outboxID,
		CreatedAt:    now,
		TraceCarrier: carrier,
		Event: dto.AvatarUploadedEvent{
			AvatarID: "avatar-1",
			UserID:   "user-1",
			S3Key:    "user-1/avatar-1/original",
		},
	}))

	events, err := repo.MarkPublishing(context.Background(), 10, now)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, carrier, events[0].TraceCarrier)
}
