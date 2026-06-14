package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCompress(t *testing.T) {
	assert.True(t, shouldCompress("application/json; charset=utf-8"))
	assert.True(t, shouldCompress("text/html"))
	assert.False(t, shouldCompress("image/png"))
	assert.False(t, shouldCompress(""))
}

func TestCompress_responseGzip(t *testing.T) {
	comp, err := NewGzipCompressor(gzip.DefaultCompression)
	require.NoError(t, err)

	e := echo.New()
	e.Use(Compress(logger.NewNopLogger(), comp))
	e.GET("/", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "application/json", []byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gr.Close()
	out, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(out))
}

func TestCompress_requestGzip(t *testing.T) {
	comp, err := NewGzipCompressor(gzip.DefaultCompression)
	require.NoError(t, err)

	var got string
	e := echo.New()
	e.Use(Compress(logger.NewNopLogger(), comp))
	e.POST("/", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		got = string(b)
		return c.NoContent(http.StatusOK)
	})

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write([]byte(`{"in":1}`))
	_ = zw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &gz)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"in":1}`, got)
}
