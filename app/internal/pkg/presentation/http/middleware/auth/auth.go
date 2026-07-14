// Package auth provides HTTP authentication middleware.
package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// UserIDHeader is the request header that carries caller identity.
const UserIDHeader = "X-User-ID"

// UserIDKey is the Echo context key for request user id.
const UserIDKey = "user_id"

// UserID returns request user id from Echo context.
func UserID(c echo.Context) (string, bool) {
	v, ok := c.Get(UserIDKey).(string)
	if !ok {
		return "", false
	}
	return v, v != ""
}

// AuthMiddleware validates X-User-ID header for protected endpoints.
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := strings.TrimSpace(c.Request().Header.Get(UserIDHeader))
			if userID == "" {
				return c.NoContent(http.StatusUnauthorized)
			}
			c.Set(UserIDKey, userID)
			return next(c)
		}
	}
}
