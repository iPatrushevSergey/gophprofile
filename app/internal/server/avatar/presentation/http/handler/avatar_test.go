package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pkgportmocks "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	appdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
	presdto "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/http/dto"
	presport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/presentation/port"
)

type stubUseCase[In, Out any] struct {
	out Out
	err error
}

func (s stubUseCase[In, Out]) Execute(_ context.Context, _ In) (Out, error) {
	return s.out, s.err
}

type stubAvatarUseCases struct {
	upload                 appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput]
	get                    appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput]
	getMetadata            appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput]
	listByUser             appport.UseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput]
	delete                 appport.UseCase[appdto.DeleteAvatarInput, struct{}]
	health                 appport.UseCase[struct{}, appdto.HealthCheckOutput]
	expireUploading        appport.UseCase[struct{}, struct{}]
	publishOutbox          appport.UseCase[struct{}, struct{}]
	collectPeriodicMetrics appport.UseCase[struct{}, struct{}]
}

func (s stubAvatarUseCases) UploadUseCase() appport.UseCase[appdto.UploadAvatarInput, appdto.UploadAvatarOutput] {
	return s.upload
}

func (s stubAvatarUseCases) GetUseCase() appport.UseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput] {
	return s.get
}

func (s stubAvatarUseCases) GetMetadataUseCase() appport.UseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput] {
	return s.getMetadata
}

func (s stubAvatarUseCases) ListByUserUseCase() appport.UseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput] {
	return s.listByUser
}

func (s stubAvatarUseCases) DeleteUseCase() appport.UseCase[appdto.DeleteAvatarInput, struct{}] {
	return s.delete
}

func (s stubAvatarUseCases) HealthUseCase() appport.UseCase[struct{}, appdto.HealthCheckOutput] {
	return s.health
}

func (s stubAvatarUseCases) ExpireUploadingAvatarsUseCase() appport.UseCase[struct{}, struct{}] {
	return s.expireUploading
}

func (s stubAvatarUseCases) PublishPendingOutboxEventsUseCase() appport.UseCase[struct{}, struct{}] {
	return s.publishOutbox
}

func (s stubAvatarUseCases) CollectPeriodicMetricsUseCase() appport.UseCase[struct{}, struct{}] {
	return s.collectPeriodicMetrics
}

var _ presport.AvatarUseCases = stubAvatarUseCases{}

func newTestAvatarHandler(t *testing.T, uc stubAvatarUseCases) *AvatarHandler {
	t.Helper()
	ctrl := gomock.NewController(t)
	log := pkgportmocks.NewMockLogger(ctrl)
	log.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return NewAvatarHandler(uc, log)
}

func TestAvatarHandler_Health(t *testing.T) {
	uc := stubAvatarUseCases{
		health: stubUseCase[struct{}, appdto.HealthCheckOutput]{
			out: appdto.HealthCheckOutput{
				Status:   "ok",
				Database: "ok",
				Storage:  "ok",
				Broker:   "ok",
			},
		},
	}
	h := newTestAvatarHandler(t, uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, h.Health(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp presdto.HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp.Status)
}

func TestAvatarHandler_Health_degraded(t *testing.T) {
	uc := stubAvatarUseCases{
		health: stubUseCase[struct{}, appdto.HealthCheckOutput]{
			out: appdto.HealthCheckOutput{
				Status:   "error",
				Database: "error",
				Storage:  "ok",
				Broker:   "ok",
			},
		},
	}
	h := newTestAvatarHandler(t, uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, h.Health(c))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestAvatarHandler_GetMetadata(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	uc := stubAvatarUseCases{
		getMetadata: stubUseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput]{
			out: appdto.AvatarMetadataOutput{
				ID:        "avatar-1",
				UserID:    "user-1",
				FileName:  "avatar.png",
				MimeType:  "image/png",
				SizeBytes: 100,
				Width:     64,
				Height:    64,
				ThumbnailS3Keys: map[vo.ThumbnailSize]map[vo.OutputFormat]string{
					vo.ThumbnailSize100: {vo.OutputFormatJPEG: "user-1/avatar-1/100x100/jpeg"},
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	h := newTestAvatarHandler(t, uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/avatar-1/metadata", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("avatar_id")
	c.SetParamValues("avatar-1")

	require.NoError(t, h.GetMetadata(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp presdto.AvatarMetadataResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "avatar-1", resp.ID)
	assert.Len(t, resp.Thumbnails, 1)
}

func TestAvatarHandler_GetMetadata_notFound(t *testing.T) {
	uc := stubAvatarUseCases{
		getMetadata: stubUseCase[appdto.GetAvatarMetadataInput, appdto.AvatarMetadataOutput]{
			err: application.ErrNotFound,
		},
	}
	h := newTestAvatarHandler(t, uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/missing/metadata", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("avatar_id")
	c.SetParamValues("missing")

	require.NoError(t, h.GetMetadata(c))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAvatarHandler_Get_badSize(t *testing.T) {
	h := newTestAvatarHandler(t, stubAvatarUseCases{})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/avatar-1?size=bad", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("avatar_id")
	c.SetParamValues("avatar-1")

	require.NoError(t, h.Get(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAvatarHandler_Get_success(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	uc := stubAvatarUseCases{
		get: stubUseCase[appdto.GetAvatarInput, appdto.GetAvatarOutput]{
			out: appdto.GetAvatarOutput{
				AvatarID:  "avatar-1",
				MimeType:  "image/jpeg",
				Content:   []byte("jpeg"),
				UpdatedAt: now,
			},
		},
	}
	h := newTestAvatarHandler(t, uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/avatars/avatar-1?size=100x100", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("avatar_id")
	c.SetParamValues("avatar-1")

	require.NoError(t, h.Get(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
}

func TestAvatarHandler_ListByUser(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	uc := stubAvatarUseCases{
		listByUser: stubUseCase[appdto.ListUserAvatarsInput, appdto.ListUserAvatarsOutput]{
			out: appdto.ListUserAvatarsOutput{
				Items: []appdto.AvatarMetadataOutput{{
					ID:        "avatar-1",
					UserID:    "user-1",
					FileName:  "avatar.png",
					MimeType:  "image/png",
					SizeBytes: 100,
					CreatedAt: now,
					UpdatedAt: now,
				}},
			},
		},
	}
	h := newTestAvatarHandler(t, uc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/user-1/avatars", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("user_id")
	c.SetParamValues("user-1")

	require.NoError(t, h.ListByUser(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAvatarHandler_Delete_unauthorized(t *testing.T) {
	h := newTestAvatarHandler(t, stubAvatarUseCases{})

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/avatar-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("avatar_id")
	c.SetParamValues("avatar-1")

	require.NoError(t, h.Delete(c))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
