package logger

import (
	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"go.opentelemetry.io/otel/trace"
)

// DefaultLogFormatter implements standard logging logic.
type DefaultLogFormatter struct{}

// Log writes a default log entry.
func (f *DefaultLogFormatter) Log(log pkgport.Logger, p LogParams) {
	args := []any{
		"uri", p.Ctx.Request().RequestURI,
		"method", p.Ctx.Request().Method,
		"duration", p.Duration,
		"status", p.Ctx.Response().Status,
		"size", p.Ctx.Response().Size,
	}

	if spanContext := trace.SpanFromContext(p.Ctx.Request().Context()).SpanContext(); spanContext.IsValid() {
		args = append(args, "trace_id", spanContext.TraceID().String(), "span_id", spanContext.SpanID().String())
	}

	log.Info("HTTP request", args...)

	if len(p.RequestBody) == 0 && p.ResponseBody.Len() == 0 {
		return
	}

	log.Debug("HTTP request/response body",
		"request_body", string(p.RequestBody),
		"response_body", p.ResponseBody.String(),
	)
}
