package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	metricsadapter "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/metrics"
	pkgportmocks "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestUploadAvatar_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		storage := portmocks.NewMockAvatarStorage(ctrl)
		outbox := portmocks.NewMockOutboxWriter(ctrl)
		transactor := portmocks.NewMockTransactor(ctrl)
		idGen := portmocks.NewMockIDGenerator(ctrl)
		clock := portmocks.NewMockClock(ctrl)
		tracer := pkgportmocks.NewMockTracer(ctrl)
		span := pkgportmocks.NewMockSpan(ctrl)
		span.EXPECT().Fail(gomock.Any())
		span.EXPECT().End()
		tracer.EXPECT().Start(ctx, gomock.Any()).Return(ctx, span)
		span.EXPECT().AddAttributes(gomock.Any())

		idGen.EXPECT().NewID().Return("avatar-1", nil)
		idGen.EXPECT().NewID().Return("outbox-1", nil)
		clock.EXPECT().Now().Return(now).AnyTimes()
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
		)
		tracer.EXPECT().ContextToMap(ctx).Return(map[string]string{"traceparent": "00-test"})
		writer.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, avatar *entity.Avatar) error {
			assert.Equal(t, "avatar-1", avatar.ID)
			assert.Equal(t, "user-1", avatar.UserID)
			assert.Equal(t, vo.UploadStatusUploading, avatar.UploadStatus)
			assert.Equal(t, vo.ProcessingStatusPending, avatar.ProcessingStatus)
			return nil
		})
		storage.EXPECT().Put(ctx, "user-1/avatar-1/original", []byte("image"), "image/png").Return(nil)
		writer.EXPECT().MarkUploadCompleted(ctx, "avatar-1", now).Return(nil)
		outbox.EXPECT().CreateUploaded(ctx, dto.OutboxUploadedCreate{
			ID:           "outbox-1",
			CreatedAt:    now,
			TraceCarrier: map[string]string{"traceparent": "00-test"},
			Event: dto.AvatarUploadedEvent{
				AvatarID: "avatar-1",
				UserID:   "user-1",
				S3Key:    "user-1/avatar-1/original",
			},
		}).Return(nil)

		uc := NewUploadAvatar(writer, storage, outbox, transactor, idGen, clock, tracer, metricsadapter.NewNopMetrics())
		out, err := uc.Execute(ctx, dto.UploadAvatarInput{
			UserID:   "user-1",
			FileName: "avatar.png",
			MimeType: "image/png",
			Content:  []byte("image"),
		})
		require.NoError(t, err)
		assert.Equal(t, "avatar-1", out.ID)
		assert.Equal(t, vo.UploadStatusCompleted, out.UploadStatus)
		assert.Equal(t, vo.ProcessingStatusPending, out.ProcessingStatus)
	})

	t.Run("invalidMimeType", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		clock := portmocks.NewMockClock(ctrl)
		clock.EXPECT().Now().Return(now).AnyTimes()
		uc := NewUploadAvatar(
			portmocks.NewMockAvatarWriter(ctrl),
			portmocks.NewMockAvatarStorage(ctrl),
			portmocks.NewMockOutboxWriter(ctrl),
			portmocks.NewMockTransactor(ctrl),
			portmocks.NewMockIDGenerator(ctrl),
			clock,
			pkgportmocks.NewMockTracer(ctrl),
			metricsadapter.NewNopMetrics(),
		)

		_, err := uc.Execute(ctx, dto.UploadAvatarInput{
			UserID:   "user-1",
			FileName: "avatar.png",
			MimeType: "text/plain",
			Content:  []byte("image"),
		})
		assert.ErrorIs(t, err, application.ErrBadInput)
	})
}
