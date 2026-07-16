package otel

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil"
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

const scopeName = "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/otel"

// Default histogram bucket boundaries in seconds; matches prometheus.DefBuckets.
var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// Version returns the instrumentation library version.
func Version() string {
	if v := strings.TrimSpace(apputil.Version); v != "" {
		return v
	}

	return "0.0.0"
}

// Metrics implements port.Metrics with OpenTelemetry instruments.
type Metrics struct {
	uploadsTotal        metric.Int64Counter
	uploadDuration      metric.Float64Histogram
	processingTotal     metric.Int64Counter
	processingDuration  metric.Float64Histogram
	deletesTotal        metric.Int64Counter
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	storageBytes        metric.Float64Gauge
	outboxPending       metric.Int64Gauge
	dbPoolTotalConns    metric.Int64Gauge
	dbPoolIdleConns     metric.Int64Gauge
	dbPoolAcquiredConns metric.Int64Gauge
}

var _ pkgport.Metrics = (*Metrics)(nil)

// NewMetrics creates an OpenTelemetry metrics adapter using the global MeterProvider.
func NewMetrics() (*Metrics, error) {
	meter := otel.Meter(scopeName, metric.WithInstrumentationVersion(Version()))

	uploadsTotal, err := meter.Int64Counter(
		"avatars_uploads",
		metric.WithDescription("Total number of avatar uploads."),
	)
	if err != nil {
		return nil, err
	}

	uploadDuration, err := meter.Float64Histogram(
		"avatars_upload_duration_seconds",
		metric.WithDescription("Avatar upload duration in seconds."),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		return nil, err
	}

	processingTotal, err := meter.Int64Counter(
		"avatars_processing",
		metric.WithDescription("Total number of avatar processing runs."),
	)
	if err != nil {
		return nil, err
	}

	processingDuration, err := meter.Float64Histogram(
		"avatars_processing_duration_seconds",
		metric.WithDescription("Avatar processing duration in seconds."),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		return nil, err
	}

	deletesTotal, err := meter.Int64Counter(
		"avatars_deletes",
		metric.WithDescription("Total number of avatar delete requests."),
	)
	if err != nil {
		return nil, err
	}

	httpRequestsTotal, err := meter.Int64Counter(
		"http_server_requests",
		metric.WithDescription("Total number of HTTP requests processed."),
	)
	if err != nil {
		return nil, err
	}

	httpRequestDuration, err := meter.Float64Histogram(
		"http_server_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds."),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		return nil, err
	}

	storageBytes, err := meter.Float64Gauge(
		"avatars_storage_bytes",
		metric.WithDescription("Total stored avatar bytes for active records."),
	)
	if err != nil {
		return nil, err
	}

	outboxPending, err := meter.Int64Gauge(
		"outbox_pending_messages",
		metric.WithDescription("Number of pending or publishing outbox messages."),
	)
	if err != nil {
		return nil, err
	}

	dbPoolTotalConns, err := meter.Int64Gauge(
		"db_pool_connections_total",
		metric.WithDescription("Total number of connections in the PostgreSQL pool."),
	)
	if err != nil {
		return nil, err
	}

	dbPoolIdleConns, err := meter.Int64Gauge(
		"db_pool_connections_idle",
		metric.WithDescription("Number of idle connections in the PostgreSQL pool."),
	)
	if err != nil {
		return nil, err
	}

	dbPoolAcquiredConns, err := meter.Int64Gauge(
		"db_pool_connections_acquired",
		metric.WithDescription("Number of acquired connections in the PostgreSQL pool."),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		uploadsTotal:        uploadsTotal,
		uploadDuration:      uploadDuration,
		processingTotal:     processingTotal,
		processingDuration:  processingDuration,
		deletesTotal:        deletesTotal,
		httpRequestsTotal:   httpRequestsTotal,
		httpRequestDuration: httpRequestDuration,
		storageBytes:        storageBytes,
		outboxPending:       outboxPending,
		dbPoolTotalConns:    dbPoolTotalConns,
		dbPoolIdleConns:     dbPoolIdleConns,
		dbPoolAcquiredConns: dbPoolAcquiredConns,
	}, nil
}

// RecordUpload records metrics for avatar uploads.
func (m *Metrics) RecordUpload(ctx context.Context, status string, duration time.Duration) {
	attrs := metric.WithAttributes(attribute.String("status", status))
	m.uploadsTotal.Add(ctx, 1, attrs)
	m.uploadDuration.Record(ctx, duration.Seconds(), attrs)
}

// RecordProcessing records metrics for avatar processing.
func (m *Metrics) RecordProcessing(ctx context.Context, status string, duration time.Duration) {
	attrs := metric.WithAttributes(attribute.String("status", status))
	m.processingTotal.Add(ctx, 1, attrs)
	m.processingDuration.Record(ctx, duration.Seconds(), attrs)
}

// RecordDelete records metrics for avatar deletions.
func (m *Metrics) RecordDelete(ctx context.Context, status string) {
	attrs := metric.WithAttributes(attribute.String("status", status))
	m.deletesTotal.Add(ctx, 1, attrs)
}

// RecordHTTPRequest records metrics for HTTP requests.
func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, route, statusClass string, duration time.Duration) {
	requestAttrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("route", route),
		attribute.String("status_class", statusClass),
	)
	durationAttrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("route", route),
	)
	m.httpRequestsTotal.Add(ctx, 1, requestAttrs)
	m.httpRequestDuration.Record(ctx, duration.Seconds(), durationAttrs)
}

// SetStorageBytes sets the total stored avatar bytes for active records.
func (m *Metrics) SetStorageBytes(ctx context.Context, bytes int64) {
	m.storageBytes.Record(ctx, float64(bytes))
}

// SetOutboxPending sets the number of pending or publishing outbox messages.
func (m *Metrics) SetOutboxPending(ctx context.Context, count int64) {
	m.outboxPending.Record(ctx, count)
}

// ObserveDBPool observes the database pool statistics.
func (m *Metrics) ObserveDBPool(ctx context.Context, stats pkgport.DBPoolStats) {
	m.dbPoolTotalConns.Record(ctx, int64(stats.TotalConns))
	m.dbPoolIdleConns.Record(ctx, int64(stats.IdleConns))
	m.dbPoolAcquiredConns.Record(ctx, int64(stats.AcquiredConns))
}
