package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

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
