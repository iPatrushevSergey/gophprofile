package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	authmw "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/presentation/http/middleware/auth"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
	presdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/http/dto"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

// maxAvatarSize is the maximum size of an avatar in bytes.
const maxAvatarSize = 10 << 20

// AvatarHandler serves avatar HTTP endpoints.
type AvatarHandler struct {
	useCases presport.AvatarUseCases
	log      pkgport.Logger
}

// NewAvatarHandler creates an avatar HTTP handler.
func NewAvatarHandler(useCases presport.AvatarUseCases, log pkgport.Logger) *AvatarHandler {
	return &AvatarHandler{useCases: useCases, log: log}
}

// Upload handles avatar upload.
func (h *AvatarHandler) Upload(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := authmw.UserID(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
	}
	if fileHeader.Size > maxAvatarSize {
		return c.JSON(http.StatusRequestEntityTooLarge, presdto.ErrorResponse{
			Error:   "File too large",
			MaxSize: maxAvatarSize,
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.log.Error(ctx, "open upload file failed", "error", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		h.log.Error(ctx, "read upload file failed", "error", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(content) > maxAvatarSize {
		return c.JSON(http.StatusRequestEntityTooLarge, presdto.ErrorResponse{
			Error:   "File too large",
			MaxSize: maxAvatarSize,
		})
	}

	out, err := h.useCases.UploadUseCase().Execute(
		ctx,
		appdto.UploadAvatarInput{
			UserID:   userID,
			FileName: fileHeader.Filename,
			MimeType: fileHeader.Header.Get("Content-Type"),
			Content:  content,
		},
	)
	if err != nil {
		switch err {
		case application.ErrBadInput:
			return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{
				Error:   "Invalid file format",
				Details: "Supported formats: jpeg, png, webp",
			})
		default:
			h.log.Error(ctx, "upload avatar failed", "error", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	return c.JSON(http.StatusCreated, presdto.UploadAvatarResponse{
		ID:        out.ID,
		UserID:    out.UserID,
		URL:       fmt.Sprintf("/api/v1/avatars/%s", out.ID),
		Status:    string(out.ProcessingStatus),
		CreatedAt: out.CreatedAt,
	})
}

// Get handles avatar retrieval.
func (h *AvatarHandler) Get(c echo.Context) error {
	ctx := c.Request().Context()

	avatarID := c.Param("avatar_id")
	size := vo.ThumbnailSize(c.QueryParam("size"))
	if size == "" {
		size = vo.ThumbnailOriginal
	}

	if !size.Valid() {
		return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
	}

	format := vo.OutputFormat(c.QueryParam("format"))
	if !format.Valid() {
		return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
	}

	if ifNoneMatch := c.Request().Header.Get("If-None-Match"); ifNoneMatch != "" {
		meta, err := h.useCases.GetMetadataUseCase().Execute(ctx, appdto.GetAvatarMetadataInput{AvatarID: avatarID})
		if err != nil {
			switch err {
			case application.ErrBadInput:
				return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
			case application.ErrNotFound:
				return c.JSON(http.StatusNotFound, presdto.ErrorResponse{Error: "Avatar not found"})
			default:
				h.log.Error(ctx, "get avatar metadata failed", "error", err)
				return c.NoContent(http.StatusInternalServerError)
			}
		}

		if size != vo.ThumbnailOriginal {
			variants, ok := meta.ThumbnailS3Keys[size]
			if !ok || len(variants) == 0 {
				return c.JSON(http.StatusNotFound, presdto.ErrorResponse{Error: "Avatar not found"})
			}
		}

		etag := fmt.Sprintf(`"%s-%s-%d"`, meta.ID, size, meta.UpdatedAt.UTC().Unix())
		for _, part := range strings.Split(ifNoneMatch, ",") {
			if strings.TrimSpace(part) == etag {
				c.Response().Header().Set("Cache-Control", "max-age=86400")
				c.Response().Header().Set("ETag", etag)
				return c.NoContent(http.StatusNotModified)
			}
		}
	}

	out, err := h.useCases.GetUseCase().Execute(ctx, appdto.GetAvatarInput{
		AvatarID:      avatarID,
		ThumbnailSize: size,
		OutputFormat:  format,
	})
	if err != nil {
		switch err {
		case application.ErrBadInput:
			return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
		case application.ErrNotFound:
			return c.JSON(http.StatusNotFound, presdto.ErrorResponse{Error: "Avatar not found"})
		default:
			h.log.Error(ctx, "get avatar failed", "error", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	etag := fmt.Sprintf(`"%s-%s-%d"`, out.AvatarID, size, out.UpdatedAt.UTC().Unix())
	c.Response().Header().Set("Cache-Control", "max-age=86400")
	c.Response().Header().Set("ETag", etag)
	return c.Blob(http.StatusOK, out.MimeType, out.Content)
}

// GetMetadata handles avatar metadata retrieval.
func (h *AvatarHandler) GetMetadata(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.useCases.GetMetadataUseCase().Execute(
		ctx,
		appdto.GetAvatarMetadataInput{
			AvatarID: c.Param("avatar_id"),
		},
	)
	if err != nil {
		switch err {
		case application.ErrBadInput:
			return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
		case application.ErrNotFound:
			return c.JSON(http.StatusNotFound, presdto.ErrorResponse{Error: "Avatar not found"})
		default:
			h.log.Error(ctx, "get avatar metadata failed", "error", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	thumbnails := make([]presdto.ThumbnailURL, 0, len(out.ThumbnailS3Keys))
	for thumbSize := range out.ThumbnailS3Keys {
		thumbnails = append(thumbnails, presdto.ThumbnailURL{
			Size: string(thumbSize),
			URL:  fmt.Sprintf("/api/v1/avatars/%s?size=%s", out.ID, thumbSize),
		})
	}

	var dimensions *presdto.ImageDimensions
	if out.Width > 0 && out.Height > 0 {
		dimensions = &presdto.ImageDimensions{
			Width:  out.Width,
			Height: out.Height,
		}
	}

	return c.JSON(http.StatusOK, presdto.AvatarMetadataResponse{
		ID:         out.ID,
		UserID:     out.UserID,
		FileName:   out.FileName,
		MimeType:   out.MimeType,
		SizeBytes:  out.SizeBytes,
		Dimensions: dimensions,
		Thumbnails: thumbnails,
		CreatedAt:  out.CreatedAt,
		UpdatedAt:  out.UpdatedAt,
	})
}

// ListByUser handles listing avatars for a user.
func (h *AvatarHandler) ListByUser(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.useCases.ListByUserUseCase().Execute(
		ctx,
		appdto.ListUserAvatarsInput{UserID: c.Param("user_id")},
	)
	if err != nil {
		switch err {
		case application.ErrBadInput:
			return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
		default:
			h.log.Error(ctx, "list user avatars failed", "error", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	items := make([]presdto.AvatarMetadataResponse, 0, len(out.Items))
	for _, item := range out.Items {
		thumbnails := make([]presdto.ThumbnailURL, 0, len(item.ThumbnailS3Keys))
		for thumbSize := range item.ThumbnailS3Keys {
			thumbnails = append(thumbnails, presdto.ThumbnailURL{
				Size: string(thumbSize),
				URL:  fmt.Sprintf("/api/v1/avatars/%s?size=%s", item.ID, thumbSize),
			})
		}

		var dimensions *presdto.ImageDimensions
		if item.Width > 0 && item.Height > 0 {
			dimensions = &presdto.ImageDimensions{
				Width:  item.Width,
				Height: item.Height,
			}
		}

		items = append(items, presdto.AvatarMetadataResponse{
			ID:         item.ID,
			UserID:     item.UserID,
			FileName:   item.FileName,
			MimeType:   item.MimeType,
			SizeBytes:  item.SizeBytes,
			Dimensions: dimensions,
			Thumbnails: thumbnails,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, presdto.UserAvatarsResponse{Items: items})
}

// Delete handles avatar deletion.
func (h *AvatarHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := authmw.UserID(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	_, err := h.useCases.DeleteUseCase().Execute(
		ctx,
		appdto.DeleteAvatarInput{
			AvatarID:      c.Param("avatar_id"),
			RequestUserID: userID,
		},
	)
	if err != nil {
		switch err {
		case application.ErrBadInput:
			return c.JSON(http.StatusBadRequest, presdto.ErrorResponse{Error: "bad input"})
		case application.ErrNotFound:
			return c.JSON(http.StatusNotFound, presdto.ErrorResponse{Error: "Avatar not found"})
		case application.ErrForbidden:
			return c.JSON(http.StatusForbidden, presdto.ErrorResponse{
				Error:   "Forbidden",
				Details: "You can only delete your own avatars",
			})
		default:
			h.log.Error(ctx, "delete avatar failed", "error", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// Health handles health check.
func (h *AvatarHandler) Health(c echo.Context) error {
	ctx := c.Request().Context()

	out, err := h.useCases.HealthUseCase().Execute(ctx, struct{}{})
	if err != nil {
		h.log.Error(ctx, "health check failed", "error", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	status := http.StatusOK
	if out.Status != "ok" {
		status = http.StatusServiceUnavailable
	}

	return c.JSON(status, presdto.HealthResponse{
		Status:   out.Status,
		Database: out.Database,
		Storage:  out.Storage,
		Broker:   out.Broker,
	})
}
