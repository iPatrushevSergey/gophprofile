package metric

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"go.opentelemetry.io/otel"

	metricsadapter "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics"
	otelmetrics "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/otel"
)

func newTestMetrics(t *testing.T) (*otelmetrics.Metrics, *sdkmetric.ManualReader) {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)
	t.Cleanup(func() { _ = mp.Shutdown(context.Background()) })

	metrics, err := otelmetrics.NewMetrics()
	require.NoError(t, err)
	return metrics, reader
}

func TestMetricsMiddleware_recordsRequest(t *testing.T) {
	metrics, reader := newTestMetrics(t)

	e := echo.New()
	e.Use(MetricsMiddleware(metrics))
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http_server_requests" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok)
			require.Len(t, sum.DataPoints, 1)
			assert.Equal(t, int64(1), sum.DataPoints[0].Value)
			found = true
		}
	}
	require.True(t, found)
}

func TestMetricsMiddleware_usesNopMetrics(t *testing.T) {
	e := echo.New()
	e.Use(MetricsMiddleware(metricsadapter.NewNopMetrics()))
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestStatusClass(t *testing.T) {
	require.Equal(t, "2xx", statusClass(201))
	require.Equal(t, "4xx", statusClass(404))
	require.Equal(t, "5xx", statusClass(503))
}

func TestResolveStatusCode(t *testing.T) {
	newContext := func() echo.Context {
		e := echo.New()
		return e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	}

	t.Run("noErrorUsesResponseStatus", func(t *testing.T) {
		c := newContext()
		c.Response().Status = http.StatusCreated
		require.Equal(t, http.StatusCreated, resolveStatusCode(c, nil))
	})

	t.Run("httpErrorUsesCode", func(t *testing.T) {
		require.Equal(t, http.StatusNotFound, resolveStatusCode(newContext(), echo.NewHTTPError(http.StatusNotFound, "missing")))
	})

	t.Run("wrappedHTTPErrorUsesCode", func(t *testing.T) {
		err := fmt.Errorf("lookup failed: %w", echo.NewHTTPError(http.StatusForbidden, "forbidden"))
		require.Equal(t, http.StatusForbidden, resolveStatusCode(newContext(), err))
	})

	t.Run("plainErrorDefaultsTo500", func(t *testing.T) {
		require.Equal(t, http.StatusInternalServerError, resolveStatusCode(newContext(), errors.New("db down")))
	})

	t.Run("plainErrorWithWrittenStatusUsesStatus", func(t *testing.T) {
		c := newContext()
		c.Response().Status = http.StatusBadGateway
		require.Equal(t, http.StatusBadGateway, resolveStatusCode(c, errors.New("upstream failed")))
	})
}

func TestRecordHTTPRequest_durationPassed(t *testing.T) {
	metrics, reader := newTestMetrics(t)
	metrics.RecordHTTPRequest(t.Context(), http.MethodGet, "/test", "2xx", 250*time.Millisecond)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	var found bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http_server_request_duration_seconds" {
				continue
			}
			hist, ok := m.Data.(metricdata.Histogram[float64])
			require.True(t, ok)
			require.NotEmpty(t, hist.DataPoints)
			assert.Equal(t, uint64(1), hist.DataPoints[0].Count)
			found = true
		}
	}
	require.True(t, found)
}
