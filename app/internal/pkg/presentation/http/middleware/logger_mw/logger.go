// Package logger_mw provides HTTP request logging middleware.
package logger_mw

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/labstack/echo/v4"
)

// LogFormatter abstraction for logging HTTP request details.
type LogFormatter interface {
	Log(log port.Logger, params LogParams)
}

// LogParams contains data available for logging.
type LogParams struct {
	Ctx          echo.Context
	Duration     time.Duration
	RequestBody  []byte
	ResponseBody *bytes.Buffer
}

// Logger is HTTP request logging middleware with injected formatter.
func Logger(log port.Logger, formatter LogFormatter) echo.MiddlewareFunc {
	if formatter == nil {
		formatter = &DefaultLogFormatter{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			var requestBody []byte
			if c.Request().Body != nil && c.Request().ContentLength != 0 {
				var readErr error
				requestBody, readErr = io.ReadAll(c.Request().Body)
				if readErr != nil {
					log.Error("failed to read request body", "error", readErr)
					return c.NoContent(http.StatusBadRequest)
				}
				c.Request().Body = io.NopCloser(bytes.NewReader(requestBody))
			}

			res := c.Response()
			w := &responseBodyWriter{body: bytes.NewBuffer(nil), ResponseWriter: res.Writer}
			res.Writer = w

			err := next(c)

			formatter.Log(log, LogParams{
				Ctx:          c,
				Duration:     time.Since(start),
				RequestBody:  requestBody,
				ResponseBody: w.body,
			})

			return err
		}
	}
}

type responseBodyWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r responseBodyWriter) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return io.WriteString(r.ResponseWriter, s)
}
