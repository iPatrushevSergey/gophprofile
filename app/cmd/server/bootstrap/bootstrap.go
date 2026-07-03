// Package bootstrap is the composition root for the GophProfile HTTP server.
package bootstrap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/retry"
	oteltelemetry "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/telemetry/otel"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	avatarclock "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/clock"
	avatargenerator "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/generator"
	avatarminio "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/minio"
	avatarpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres"
	avatarrmq "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/rabbitmq"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/config"
)

// Run loads configuration, wires dependencies and runs the application.
func Run() error {
	// Load config.
	cfg, err := config.LoadConfig()
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize logger.
	var _ pkgport.Logger = (*logger.ZapLogger)(nil)
	log, err := logger.NewZapLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	// Initialize OpenTelemetry.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var telemetryShutdown func(context.Context) error
	if cfg.Telemetry.Enabled {
		telCtx := context.Background()

		exporterOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Telemetry.OTLPEndpoint),
		}
		if cfg.Telemetry.OTLPInsecure {
			exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
		}

		exporter, err := otlptracegrpc.New(telCtx, exporterOpts...)
		if err != nil {
			return fmt.Errorf("create otlp trace exporter: %w", err)
		}

		res, err := resource.New(telCtx,
			resource.WithAttributes(
				semconv.ServiceName(cfg.Telemetry.ServiceName),
				semconv.ServiceVersion(apputil.Version),
				semconv.DeploymentEnvironmentName(cfg.Telemetry.Environment),
			),
			resource.WithFromEnv(),
			resource.WithProcess(),
			resource.WithOS(),
		)
		if err != nil {
			return fmt.Errorf("create otel resource: %w", err)
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Telemetry.SampleRatio))),
		)
		otel.SetTracerProvider(tp)
		telemetryShutdown = tp.Shutdown
	}

	// Initialize tracer.
	tracer := oteltelemetry.NewTracer()

	// Log server startup details.
	log.Info("starting gophprofile server",
		"address", cfg.Server.Address,
		"database_configured", cfg.DB.Pool.URI != "",
		"minio_configured", cfg.MinIO.Enabled(),
		"broker_configured", cfg.Broker.Enabled(),
		"tls_configured", cfg.Server.TLSEnabled(),
		"telemetry_enabled", cfg.Telemetry.Enabled,
	)

	// Initialize database pool.
	if cfg.DB.Pool.URI == "" {
		return fmt.Errorf("database: uri is required")
	}
	pool, err := postgres.NewPool(context.Background(), cfg.DB.Pool)
	if err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	defer pool.Close()
	log.Info("database connected")

	// Initialize retry options.
	retryOpts := []retry.RetryOption{retry.WithMaxRetries(cfg.DB.Retry.MaxRetries)}
	if cfg.DB.Retry.MaxRetries > 0 {
		retryOpts = append(retryOpts, retry.WithExponentialBackoff(cfg.DB.Retry.BaseDelay, cfg.DB.Retry.MaxDelay))
	}

	// Initialize transactor.
	transactor := postgres.NewTransactor(pool, retryOpts...)

	// Build options for the use case factory.
	useCaseOpts := []apputil.Option[globalUseCasesParams]{
		WithTransactor(transactor),
		WithAvatarRepo(avatarpostgres.NewAvatarRepository(transactor)),
		WithOutboxRepo(avatarpostgres.NewOutboxRepository(transactor)),
		WithTracer(tracer),
		WithIDGenerator(avatargenerator.NewIDGenerator()),
		WithClock(avatarclock.NewRealClock()),
		WithLogger(log),
		WithOutboxBatchSize(cfg.Broker.OutboxBatchSize),
		WithOutboxPublishingTimeout(cfg.Broker.OutboxPublishingTimeout),
		WithUploadReservationTTL(cfg.MinIO.UploadReservationTTL),
	}

	// Initialize MinIO avatar storage.
	if !cfg.MinIO.Enabled() {
		return fmt.Errorf("minio: endpoint, access_key and secret_key are required")
	}
	avatarStorage, err := avatarminio.NewAvatarStorage(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("minio avatar storage: %w", err)
	}
	log.Info("minio avatar storage configured", "endpoint", cfg.MinIO.Endpoint, "bucket", avatarStorage.Bucket())
	useCaseOpts = append(useCaseOpts, WithAvatarStorage(avatarStorage))

	// Initialize RabbitMQ event publisher.
	eventPublisher, err := avatarrmq.NewPublisher(cfg.Broker, tracer)
	if err != nil {
		return fmt.Errorf("rabbitmq event publisher: %w", err)
	}
	if cfg.Broker.Enabled() {
		log.Info("rabbitmq event publisher configured", "exchange", cfg.Broker.Exchange)
	}
	useCaseOpts = append(useCaseOpts, WithEventPublisher(eventPublisher))

	// Build use case factory.
	useCases := NewGlobalUseCases(useCaseOpts...)

	// Initialize router.
	router, err := NewGlobalRouter(useCases, log, cfg)
	if err != nil {
		return fmt.Errorf("router: %w", err)
	}

	// Initialize HTTP server.
	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router,
	}

	// Initialize TLS configuration.
	cert, key := cfg.Server.CertFile, cfg.Server.KeyFile
	if cert != "" || key != "" {
		if cert == "" || key == "" {
			return fmt.Errorf("tls: both cert_file and key_file are required")
		}
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// Initialize application.
	app := &App{
		Server:            srv,
		Log:               log,
		TelemetryShutdown: telemetryShutdown,
		ShutdownTimeout:   cfg.Server.ShutdownTimeout,
		TLSCertFile:       cert,
		TLSKeyFile:        key,
		UseCases:          useCases,
		Tracer:            tracer,
	}
	if cfg.MinIO.Enabled() {
		app.UploadGCInterval = cfg.MinIO.UploadGCInterval
	}
	if cfg.Broker.Enabled() {
		app.OutboxPublishInterval = cfg.Broker.PublishInterval
	}

	// Start application.
	app.Start()

	// Wait for application to stop.
	return app.Stop()
}
