package logger

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	"github.com/labstack/echo/v4"
)

type recordingLogger struct {
	lastInfo []any
}

func (l *recordingLogger) Debug(msg string, args ...any) {}
func (l *recordingLogger) Info(msg string, args ...any) {
	if msg == "HTTP request" {
		l.lastInfo = append([]any{}, args...)
	}
}
func (l *recordingLogger) Warn(msg string, args ...any)  {}
func (l *recordingLogger) Error(msg string, args ...any) {}
func (l *recordingLogger) Sync() error                   { return nil }

func TestDefaultLogFormatter_Log(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	f := DefaultLogFormatter{}
	f.Log(logger.NewNopLogger(), LogParams{
		Ctx:          c,
		Duration:     time.Millisecond,
		RequestBody:  []byte("req"),
		ResponseBody: bytes.NewBufferString("resp"),
	})
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDefaultLogFormatter_Log_includesTraceFields(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	ctx, span := tp.Tracer("test").Start(context.Background(), "get")
	defer span.End()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	log := &recordingLogger{}
	f := DefaultLogFormatter{}
	f.Log(log, LogParams{
		Ctx:          c,
		Duration:     time.Millisecond,
		ResponseBody: bytes.NewBuffer(nil),
	})

	require.NotEmpty(t, log.lastInfo)
	assert.Contains(t, log.lastInfo, "trace_id")
	assert.Contains(t, log.lastInfo, "span_id")
	traceIDIdx := indexOf(log.lastInfo, "trace_id")
	require.NotZero(t, traceIDIdx)
	assert.NotEmpty(t, log.lastInfo[traceIDIdx+1])
}

func indexOf(args []any, key string) int {
	for i := 0; i+1 < len(args); i += 2 {
		if k, ok := args[i].(string); ok && k == key {
			return i
		}
	}
	return -1
}
