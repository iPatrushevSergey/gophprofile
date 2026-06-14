package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	e := echo.New()
	e.Use(Logger(logger.NewNopLogger(), nil))
	e.POST("/", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "text/plain", []byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("in")))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}
