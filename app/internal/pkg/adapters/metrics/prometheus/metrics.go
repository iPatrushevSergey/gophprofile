package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// Metrics implements port.Metrics with Prometheus instruments.
type Metrics struct {
	registry *prometheus.Registry

	uploadsTotal        *prometheus.CounterVec
	uploadDuration      *prometheus.HistogramVec
	processingTotal     *prometheus.CounterVec
	processingDuration  *prometheus.HistogramVec
	deletesTotal        *prometheus.CounterVec
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	storageBytes        prometheus.Gauge
	outboxPending       prometheus.Gauge
	dbPoolTotalConns    prometheus.Gauge
	dbPoolIdleConns     prometheus.Gauge
	dbPoolAcquiredConns prometheus.Gauge
}

var _ pkgport.Metrics = (*Metrics)(nil)

// NewMetrics creates a Prometheus metrics adapter with a dedicated registry.
func NewMetrics() (*Metrics, error) {
	reg := prometheus.NewRegistry()

	m := &Metrics{
		registry: reg,
		uploadsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "avatars_uploads_total",
			Help: "Total number of avatar uploads.",
		}, []string{"status"}),
		uploadDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "avatars_upload_duration_seconds",
			Help:    "Avatar upload duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"status"}),
		processingTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "avatars_processing_total",
			Help: "Total number of avatar processing runs.",
		}, []string{"status"}),
		processingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "avatars_processing_duration_seconds",
			Help:    "Avatar processing duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"status"}),
		deletesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "avatars_deletes_total",
			Help: "Total number of avatar delete requests.",
		}, []string{"status"}),
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_server_requests_total",
			Help: "Total number of HTTP requests processed.",
		}, []string{"method", "route", "status_class"}),
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_server_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		storageBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "avatars_storage_bytes",
			Help: "Total stored avatar bytes for active records.",
		}),
		outboxPending: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outbox_pending_messages",
			Help: "Number of pending or publishing outbox messages.",
		}),
		dbPoolTotalConns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "db_pool_connections_total",
			Help: "Total number of connections in the PostgreSQL pool.",
		}),
		dbPoolIdleConns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "db_pool_connections_idle",
			Help: "Number of idle connections in the PostgreSQL pool.",
		}),
		dbPoolAcquiredConns: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "db_pool_connections_acquired",
			Help: "Number of acquired connections in the PostgreSQL pool.",
		}),
	}

	instruments := []prometheus.Collector{
		m.uploadsTotal,
		m.uploadDuration,
		m.processingTotal,
		m.processingDuration,
		m.deletesTotal,
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.storageBytes,
		m.outboxPending,
		m.dbPoolTotalConns,
		m.dbPoolIdleConns,
		m.dbPoolAcquiredConns,
	}
	for _, instrument := range instruments {
		if err := reg.Register(instrument); err != nil {
			return nil, fmt.Errorf("register prometheus metric: %w", err)
		}
	}

	return m, nil
}

// Registry returns the Prometheus registry used by this adapter.
func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

// Handler returns an HTTP handler that exposes metrics from this adapter.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// RecordUpload records metrics for avatar uploads.
func (m *Metrics) RecordUpload(_ context.Context, status string, duration time.Duration) {
	labels := prometheus.Labels{"status": status}
	m.uploadsTotal.With(labels).Inc()
	m.uploadDuration.With(labels).Observe(duration.Seconds())
}

// RecordProcessing records metrics for avatar processing.
func (m *Metrics) RecordProcessing(_ context.Context, status string, duration time.Duration) {
	labels := prometheus.Labels{"status": status}
	m.processingTotal.With(labels).Inc()
	m.processingDuration.With(labels).Observe(duration.Seconds())
}

// RecordDelete records metrics for avatar deletions.
func (m *Metrics) RecordDelete(_ context.Context, status string) {
	m.deletesTotal.With(prometheus.Labels{"status": status}).Inc()
}

// RecordHTTPRequest records metrics for HTTP requests.
func (m *Metrics) RecordHTTPRequest(_ context.Context, method, route, statusClass string, duration time.Duration) {
	requestLabels := prometheus.Labels{
		"method":       method,
		"route":        route,
		"status_class": statusClass,
	}
	durationLabels := prometheus.Labels{
		"method": method,
		"route":  route,
	}
	m.httpRequestsTotal.With(requestLabels).Inc()
	m.httpRequestDuration.With(durationLabels).Observe(duration.Seconds())
}

// SetStorageBytes sets the total stored avatar bytes for active records.
func (m *Metrics) SetStorageBytes(bytes int64) {
	m.storageBytes.Set(float64(bytes))
}

// SetOutboxPending sets the number of pending or publishing outbox messages.
func (m *Metrics) SetOutboxPending(count int64) {
	m.outboxPending.Set(float64(count))
}

// ObserveDBPool observes the database pool statistics.
func (m *Metrics) ObserveDBPool(stats pkgport.DBPoolStats) {
	m.dbPoolTotalConns.Set(float64(stats.TotalConns))
	m.dbPoolIdleConns.Set(float64(stats.IdleConns))
	m.dbPoolAcquiredConns.Set(float64(stats.AcquiredConns))
}
