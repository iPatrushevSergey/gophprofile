// Package bootstrap is the composition root for the GophProfile processor service.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/pflag"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	metricsadapter "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics"
	prommetrics "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/prometheus"
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

	// Initialize OpenTelemetry.
	var telemetryShutdown func(context.Context) error
	var otelLoggerProvider *sdklog.LoggerProvider
	var otelServiceName string
	if cfg.Telemetry.Enabled {
		telCtx := context.Background()

		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

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
			return fmt.Errorf("init telemetry: create resource: %w", err)
		}

		traceExporterOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Telemetry.OTLPEndpoint),
		}
		logExporterOpts := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(cfg.Telemetry.OTLPEndpoint),
		}
		if cfg.Telemetry.OTLPInsecure {
			traceExporterOpts = append(traceExporterOpts, otlptracegrpc.WithInsecure())
			logExporterOpts = append(logExporterOpts, otlploggrpc.WithInsecure())
		}

		traceExporter, err := otlptracegrpc.New(telCtx, traceExporterOpts...)
		if err != nil {
			return fmt.Errorf("init telemetry: create trace exporter: %w", err)
		}

		logExporter, err := otlploggrpc.New(telCtx, logExporterOpts...)
		if err != nil {
			return fmt.Errorf("init telemetry: create log exporter: %w", err)
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Telemetry.SampleRatio))),
		)
		lp := sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		)

		otel.SetTracerProvider(tp)
		global.SetLoggerProvider(lp)

		telemetryShutdown = func(ctx context.Context) error {
			var shutdownErr error
			if err := tp.Shutdown(ctx); err != nil {
				shutdownErr = fmt.Errorf("shutdown tracer provider: %w", err)
			}
			if err := lp.Shutdown(ctx); err != nil {
				if shutdownErr != nil {
					return fmt.Errorf("%w; shutdown logger provider: %v", shutdownErr, err)
				}
				return fmt.Errorf("shutdown logger provider: %w", err)
			}
			return shutdownErr
		}
		otelLoggerProvider = lp
		otelServiceName = cfg.Telemetry.ServiceName
	}

	// Initialize logger.
	var _ pkgport.Logger = (*logger.ZapLogger)(nil)
	var _ pkgport.Logger = (*logger.SlogLogger)(nil)

	var log pkgport.Logger
	switch cfg.Logger.Backend {
	case "slog":
		log, err = logger.NewSlogLogger(cfg.Logger, otelLoggerProvider, otelServiceName)
	case "zap":
		log, err = logger.NewZapLogger(cfg.Logger)
	default:
		return fmt.Errorf("init logger: unknown backend %q", cfg.Logger.Backend)
	}
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	// Initialize tracer.
	tracer := oteltelemetry.NewTracer()

	// Initialize metrics.
	var prometheusMetrics *prommetrics.Metrics
	appMetrics := pkgport.Metrics(metricsadapter.NewNopMetrics())
	if cfg.Metrics.Enabled {
		prometheusMetrics, err = prommetrics.NewMetrics()
		if err != nil {
			return fmt.Errorf("init metrics: %w", err)
		}
		appMetrics = prometheusMetrics
	}

	// Log processor startup details.
	log.Info(context.Background(), "starting gophprofile processor",
		"database_configured", cfg.DB.Pool.URI != "",
		"minio_configured", cfg.MinIO.Enabled(),
		"broker_configured", cfg.Broker.Enabled(),
		"telemetry_enabled", cfg.Telemetry.Enabled,
		"metrics_enabled", cfg.Metrics.Enabled,
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
	log.Info(context.Background(), "database connected")

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
		WithPoolStats(postgres.NewPoolStats(pool)),
		WithClock(processingclock.NewRealClock()),
		WithImageProcessor(imaging.NewProcessor()),
		WithTracer(tracer),
		WithMetrics(appMetrics),
	}

	// Initialize MinIO avatar storage.
	if !cfg.MinIO.Enabled() {
		return fmt.Errorf("minio: endpoint, access_key and secret_key are required")
	}
	avatarStorage, err := processingminio.NewAvatarStorage(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("minio avatar storage: %w", err)
	}
	log.Info(context.Background(), "minio avatar storage configured", "endpoint", cfg.MinIO.Endpoint, "bucket", cfg.MinIO.Bucket)
	useCaseOpts = append(useCaseOpts, WithAvatarStorage(avatarStorage))

	// Initialize RabbitMQ event consumer.
	if !cfg.Broker.Enabled() {
		return fmt.Errorf("broker: url is required")
	}
	eventConsumer, err := processingbroker.NewConsumer(cfg.Broker, log, tracer)
	if err != nil {
		return fmt.Errorf("rabbitmq consumer: %w", err)
	}
	log.Info(context.Background(), "rabbitmq event consumer configured",
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
		Log:                            log,
		TelemetryShutdown:              telemetryShutdown,
		ShutdownTimeout:                cfg.Worker.ShutdownTimeout,
		UseCases:                       useCases,
		EventConsumer:                  eventConsumer,
		Metrics:                        prometheusMetrics,
		MetricsAddress:                 cfg.Metrics.Address,
		MetricsEnabled:                 cfg.Metrics.Enabled,
		PeriodicMetricsCollectInterval: cfg.Metrics.CollectInterval,
	}

	// Start application.
	app.Start()

	// Wait for application to stop.
	return app.Stop()
}
