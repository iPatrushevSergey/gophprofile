package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequireUserID_missingHeader(t *testing.T) {
	e := echo.New()
	e.Use(AuthMiddleware())
	e.POST("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireUserID_validHeader(t *testing.T) {
	e := echo.New()
	e.Use(AuthMiddleware())
	e.POST("/", func(c echo.Context) error {
		userID, ok := UserID(c)
		if !ok || userID != "user-42" {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(UserIDHeader, "user-42")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireUserID_trimsHeader(t *testing.T) {
	e := echo.New()
	e.Use(AuthMiddleware())
	e.POST("/", func(c echo.Context) error {
		userID, ok := UserID(c)
		if !ok || userID != "user-42" {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(UserIDHeader, "  user-42  ")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserID_notSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, ok := UserID(c)
	assert.False(t, ok)
}
