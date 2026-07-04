package logger

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applogger "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
)

type recordingLogger struct {
	lastCtx  context.Context
	lastInfo []any
}

func (l *recordingLogger) Debug(context.Context, string, ...any) {}
func (l *recordingLogger) Info(ctx context.Context, msg string, args ...any) {
	if msg == "HTTP request" {
		l.lastCtx = ctx
		l.lastInfo = append([]any{}, args...)
	}
}
func (l *recordingLogger) Warn(context.Context, string, ...any)  {}
func (l *recordingLogger) Error(context.Context, string, ...any) {}
func (l *recordingLogger) Sync() error                           { return nil }

func TestDefaultLogFormatter_Log(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	f := &DefaultLogFormatter{}
	f.Log(applogger.NewNopLogger(), LogParams{
		Ctx:          c,
		Duration:     time.Millisecond,
		RequestBody:  []byte("req"),
		ResponseBody: bytes.NewBufferString("resp"),
	})
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDefaultLogFormatter_Log_passesRequestContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), testContextKey{}, "value")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	log := &recordingLogger{}
	(&DefaultLogFormatter{}).Log(log, LogParams{
		Ctx:          c,
		Duration:     time.Millisecond,
		ResponseBody: bytes.NewBuffer(nil),
	})

	require.NotEmpty(t, log.lastInfo)
	assert.Equal(t, ctx, log.lastCtx)
}

type testContextKey struct{}
