package metrics

import (
	"context"
	"time"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// NopMetrics discards all metric recordings.
type NopMetrics struct{}

// NewNopMetrics returns a no-op metrics implementation for tests.
func NewNopMetrics() *NopMetrics {
	return &NopMetrics{}
}

var _ pkgport.Metrics = (*NopMetrics)(nil)

// RecordUpload implements port.Metrics.
func (NopMetrics) RecordUpload(context.Context, string, time.Duration) {}

// RecordProcessing implements port.Metrics.
func (NopMetrics) RecordProcessing(context.Context, string, time.Duration) {}

// RecordDelete implements port.Metrics.
func (NopMetrics) RecordDelete(context.Context, string) {}

// RecordHTTPRequest implements port.Metrics.
func (NopMetrics) RecordHTTPRequest(context.Context, string, string, string, time.Duration) {}

// SetStorageBytes implements port.Metrics.
func (NopMetrics) SetStorageBytes(context.Context, int64) {}

// SetOutboxPending implements port.Metrics.
func (NopMetrics) SetOutboxPending(context.Context, int64) {}

// ObserveDBPool implements port.Metrics.
func (NopMetrics) ObserveDBPool(context.Context, pkgport.DBPoolStats) {}
