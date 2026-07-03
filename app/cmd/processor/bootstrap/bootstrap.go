// Package bootstrap is the composition root for the GophProfile processor service.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

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
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/config"
	processingbroker "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/broker/rabbitmq"
	processingclock "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/clock"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/imaging"
	processingminio "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/minio"
	processingpostgres "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/postgres"
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

	// Log processor startup details.
	log.Info("starting gophprofile processor",
		"database_configured", cfg.DB.Pool.URI != "",
		"minio_configured", cfg.MinIO.Enabled(),
		"broker_configured", cfg.Broker.Enabled(),
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
		WithAvatarRepo(processingpostgres.NewAvatarRepository(transactor)),
		WithClock(processingclock.NewRealClock()),
		WithImageProcessor(imaging.NewProcessor()),
		WithTracer(tracer),
	}

	// Initialize MinIO avatar storage.
	if !cfg.MinIO.Enabled() {
		return fmt.Errorf("minio: endpoint, access_key and secret_key are required")
	}
	avatarStorage, err := processingminio.NewAvatarStorage(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("minio avatar storage: %w", err)
	}
	log.Info("minio avatar storage configured", "endpoint", cfg.MinIO.Endpoint, "bucket", cfg.MinIO.Bucket)
	useCaseOpts = append(useCaseOpts, WithAvatarStorage(avatarStorage))

	// Initialize RabbitMQ event consumer.
	if !cfg.Broker.Enabled() {
		return fmt.Errorf("broker: url is required")
	}
	eventConsumer, err := processingbroker.NewConsumer(cfg.Broker, log, tracer)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: %w", err)
	}
	log.Info("rabbitmq event consumer configured",
		"exchange", cfg.Broker.Exchange,
		"queue", cfg.Broker.Queue,
		"dead_letter_exchange", cfg.Broker.DeadLetterExchange,
		"dead_letter_queue", cfg.Broker.DeadLetterQueue,
	)
	useCaseOpts = append(useCaseOpts, WithEventConsumer(eventConsumer))

	// Build use case factory.
	useCases := NewGlobalUseCases(useCaseOpts...)

	// Initialize application.
	app := &App{
		Log:               log,
		TelemetryShutdown: telemetryShutdown,
		ShutdownTimeout:   cfg.Worker.ShutdownTimeout,
		UseCases:          useCases,
		EventConsumer:     eventConsumer,
	}

	// Start application.
	app.Start()

	// Wait for application to stop.
	return app.Stop()
}
