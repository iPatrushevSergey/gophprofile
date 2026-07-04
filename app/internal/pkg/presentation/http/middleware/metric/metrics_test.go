package metric

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	metricsadapter "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics"
	prommetrics "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics/prometheus"
)

func TestMetricsMiddleware_recordsRequest(t *testing.T) {
	metrics, err := prommetrics.NewMetrics()
	require.NoError(t, err)

	e := echo.New()
	e.Use(MetricsMiddleware(metrics))
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	families, err := metrics.Registry().Gather()
	require.NoError(t, err)

	var found bool
	for _, family := range families {
		if family.GetName() != "http_server_requests_total" {
			continue
		}
		require.Len(t, family.GetMetric(), 1)
		require.Equal(t, 1.0, family.GetMetric()[0].GetCounter().GetValue())
		found = true
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
	metrics, err := prommetrics.NewMetrics()
	require.NoError(t, err)

	metrics.RecordHTTPRequest(t.Context(), http.MethodGet, "/test", "2xx", 250*time.Millisecond)

	families, err := metrics.Registry().Gather()
	require.NoError(t, err)

	var found bool
	for _, family := range families {
		if family.GetName() != "http_server_request_duration_seconds" {
			continue
		}
		require.NotEmpty(t, family.GetMetric())
		require.Equal(t, uint64(1), family.GetMetric()[0].GetHistogram().GetSampleCount())
		found = true
	}
	require.True(t, found)
}
