package port

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=metrics.go -destination=mocks/mock_metrics.go -package=mocks

import (
	"context"
	"errors"
	"time"
)

// Business metric status labels (bounded cardinality).
const (
	MetricStatusSuccess         = "success"
	MetricStatusError           = "error"
	MetricStatusValidationError = "validation_error"
)

// DBPoolStats holds PostgreSQL pool connection statistics.
type DBPoolStats struct {
	TotalConns    int32
	IdleConns     int32
	AcquiredConns int32
}

// Metrics records application and infrastructure metrics.
type Metrics interface {
	RecordUpload(ctx context.Context, status string, duration time.Duration)
	RecordProcessing(ctx context.Context, status string, duration time.Duration)
	RecordDelete(ctx context.Context, status string)
	RecordHTTPRequest(ctx context.Context, method, route, statusClass string, duration time.Duration)
	SetStorageBytes(bytes int64)
	SetOutboxPending(count int64)
	ObserveDBPool(stats DBPoolStats)
}

// MetricStatus maps an execution error to a bounded business metric status label.
func MetricStatus(err, validationErr error) string {
	if err == nil {
		return MetricStatusSuccess
	}
	if validationErr != nil && errors.Is(err, validationErr) {
		return MetricStatusValidationError
	}
	return MetricStatusError
}
