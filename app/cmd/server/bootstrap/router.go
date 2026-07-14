package bootstrap

import (
	"compress/gzip"
	"fmt"

	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	zaplogger "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/logger"
	authmw "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/presentation/http/middleware/auth"
	"github.com/iPatrushevSergey/gophprofile/app/internal/pkg/presentation/http/middleware/compression"
	mwlogger "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/presentation/http/middleware/logger"
	avatarrouter "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/http/router"
)

// NewGlobalRouter composes global middleware and module routers.
// Auth middleware applies only to routes registered inside the protected group.
func NewGlobalRouter(useCases GlobalUseCases, log *zaplogger.ZapLogger) (*echo.Echo, error) {
	r := echo.New()

	r.Use(middleware.Recover())

	gzipCompressor, err := compression.NewGzipCompressor(gzip.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("gzip compressor: %w", err)
	}
	r.Use(compression.CompressMiddleware(log, gzipCompressor))

	r.Use(mwlogger.LoggerMiddleware(log, nil))

	pprof.Register(r)

	public := r.Group("/api/v1")

	protected := r.Group("/api/v1")
	protected.Use(authmw.AuthMiddleware())

	avatarrouter.RegisterAvatarRoutes(r, public, protected, useCases, log)

	r.GET("/web", func(c echo.Context) error {
		return c.File("web/static/index.html")
	})

	return r, nil
}
