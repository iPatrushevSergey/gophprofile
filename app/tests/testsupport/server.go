//go:build e2e || component || contract

// Package testsupport provides shared helpers for GophProfile e2e tests.
package testsupport

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	serverbootstrap "github.com/iPatrushevSergey/gophprofile/app/cmd/server/bootstrap"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	metricsadapter "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics"
	otelmetrics "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/otel"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/retry"
	oteltelemetry "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/testutil"
	avatarclock "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/clock"
	avatargenerator "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/generator"
	avatarpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/config"
)

const (
	E2EUserID                  = "e2e-user"
	E2EOutboxBatchSize         = 10
	E2EOutboxPublishingTimeout = time.Minute
	E2EUploadReservationTTL    = 30 * time.Minute
)

// TestServer holds a running e2e HTTP server and its wired dependencies.
type TestServer struct {
	Server   *httptest.Server
	UseCases serverbootstrap.GlobalUseCases
}

// NewTestServer boots Postgres + MinIO + RabbitMQ, wires server use cases and returns an httptest server.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	pool := testutil.SetupPostgres(t)
	avatarStorage := testutil.SetupMinIO(t)
	eventPublisher := testutil.SetupRabbitMQ(t)

	transactor := postgres.NewTransactor(pool,
		retry.WithMaxRetries(1),
		retry.WithExponentialBackoff(50*time.Millisecond, 200*time.Millisecond),
	)

	log, err := logger.NewZapLogger(logger.Config{Level: "error", Backend: "zap", Format: "json"})
	require.NoError(t, err)

	useCases := serverbootstrap.NewGlobalUseCases(
		serverbootstrap.WithTransactor(transactor),
		serverbootstrap.WithAvatarRepo(avatarpostgres.NewAvatarRepository(transactor)),
		serverbootstrap.WithOutboxRepo(avatarpostgres.NewOutboxRepository(transactor)),
		serverbootstrap.WithAvatarStorage(avatarStorage),
		serverbootstrap.WithEventPublisher(eventPublisher),
		serverbootstrap.WithTracer(oteltelemetry.NewTracer()),
		serverbootstrap.WithMetrics(metricsadapter.NewNopMetrics()),
		serverbootstrap.WithIDGenerator(avatargenerator.NewIDGenerator()),
		serverbootstrap.WithClock(avatarclock.NewRealClock()),
		serverbootstrap.WithLogger(log),
		serverbootstrap.WithOutboxBatchSize(E2EOutboxBatchSize),
		serverbootstrap.WithOutboxPublishingTimeout(E2EOutboxPublishingTimeout),
		serverbootstrap.WithUploadReservationTTL(E2EUploadReservationTTL),
	)

	router, err := serverbootstrap.NewGlobalRouter(useCases, log, config.Config{
		Telemetry: config.Telemetry{ServiceName: "gophprofile-server"},
		Metrics:   otelmetrics.Config{Enabled: false},
	}, nil)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return &TestServer{Server: srv, UseCases: useCases}
}

// APIBase returns the server base URL for HTTP clients.
func (s *TestServer) APIBase() string {
	return s.Server.URL + "/api/v1"
}
