package logger

import pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"

// DefaultLogFormatter implements standard logging logic.
type DefaultLogFormatter struct{}

// Log writes a default log entry.
func (f *DefaultLogFormatter) Log(log pkgport.Logger, p LogParams) {
	log.Info("HTTP request",
		"uri", p.Ctx.Request().RequestURI,
		"method", p.Ctx.Request().Method,
		"duration", p.Duration,
		"status", p.Ctx.Response().Status,
		"size", p.Ctx.Response().Size,
	)
	log.Debug("HTTP request/response body",
		"request_body", string(p.RequestBody),
		"response_body", p.ResponseBody.String(),
	)
}
