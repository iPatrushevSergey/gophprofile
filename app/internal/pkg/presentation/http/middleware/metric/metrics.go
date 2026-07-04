// Package metric provides HTTP RED metrics middleware.
package metric

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
)

// MetricsMiddleware records HTTP RED metrics via the metrics port.
func MetricsMiddleware(metrics pkgport.Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)

			route := c.Path()
			if route == "" {
				route = c.Request().URL.Path
			}

			metrics.RecordHTTPRequest(
				c.Request().Context(),
				c.Request().Method,
				route,
				statusClass(resolveStatusCode(c, err)),
				time.Since(start),
			)

			return err
		}
	}
}

// resolveStatusCode resolves the status code from the echo context and error.
func resolveStatusCode(c echo.Context, err error) int {
	status := c.Response().Status
	if err == nil {
		return status
	}

	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		status = httpError.Code
	}
	if status == 0 || status == http.StatusOK {
		status = http.StatusInternalServerError
	}
	return status
}

// statusClass returns the status class for the given code.
func statusClass(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	case code >= 200:
		return "2xx"
	default:
		return strconv.Itoa(code)
	}
}
