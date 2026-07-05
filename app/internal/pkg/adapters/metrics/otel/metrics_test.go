package otel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"go.opentelemetry.io/otel"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

func newTestMetrics(t *testing.T) (*Metrics, *sdkmetric.ManualReader) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	metrics, err := NewMetrics()
	require.NoError(t, err)
	return metrics, reader
}

func TestMetrics_RecordUpload(t *testing.T) {
	metrics, reader := newTestMetrics(t)
	metrics.RecordUpload(t.Context(), pkgport.MetricStatusSuccess, 150*time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))
	require.NotEmpty(t, rm.ScopeMetrics)

	var counterFound bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "avatars_uploads" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok)
			require.Len(t, sum.DataPoints, 1)
			assert.Equal(t, int64(1), sum.DataPoints[0].Value)
			counterFound = true
		}
	}
	require.True(t, counterFound)
}

func TestMetrics_RecordHTTPRequest(t *testing.T) {
	metrics, reader := newTestMetrics(t)
	metrics.RecordHTTPRequest(t.Context(), "GET", "/test", "2xx", 250*time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	var counterFound bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http_server_requests" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok)
			require.Len(t, sum.DataPoints, 1)
			assert.Equal(t, int64(1), sum.DataPoints[0].Value)
			counterFound = true
		}
	}
	require.True(t, counterFound)
}
