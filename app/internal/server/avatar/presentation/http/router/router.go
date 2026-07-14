// Package router registers avatar HTTP routes.
package router

import (
	"github.com/labstack/echo/v4"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/http/handler"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

// RegisterAvatarRoutes registers avatar module HTTP routes.
func RegisterAvatarRoutes(
	r *echo.Echo,
	public, protected *echo.Group,
	useCases presport.AvatarUseCases,
	log pkgport.Logger,
) {
	h := handler.NewAvatarHandler(useCases, log)

	r.GET("/health", h.Health)

	public.GET("/avatars/:avatar_id", h.Get)
	public.GET("/avatars/:avatar_id/metadata", h.GetMetadata)
	public.GET("/users/:user_id/avatars", h.ListByUser)
	protected.POST("/avatars", h.Upload)
	protected.DELETE("/avatars/:avatar_id", h.Delete)
}
