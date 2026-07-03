package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	pkgportmocks "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	portmocks "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port/mocks"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

func TestDeleteAvatar_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	avatar := entity.NewAvatar(
		"avatar-1",
		"user-1",
		"avatar.png",
		"image/png",
		5,
		"user-1/avatar-1/original",
		vo.UploadStatusCompleted,
		now,
	)
	avatar.ThumbnailS3Keys = map[vo.ThumbnailSize]map[vo.OutputFormat]string{
		vo.ThumbnailSize100: {vo.OutputFormatJPEG: "user-1/avatar-1/100x100/jpeg"},
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		writer := portmocks.NewMockAvatarWriter(ctrl)
		outbox := portmocks.NewMockOutboxWriter(ctrl)
		transactor := portmocks.NewMockTransactor(ctrl)
		idGen := portmocks.NewMockIDGenerator(ctrl)
		clock := portmocks.NewMockClock(ctrl)
		tracer := pkgportmocks.NewMockTracer(ctrl)
		span := pkgportmocks.NewMockSpan(ctrl)
		span.EXPECT().Fail(gomock.Any())
		span.EXPECT().End()
		tracer.EXPECT().Start(ctx, gomock.Any()).Return(ctx, span)

		reader.EXPECT().FindByID(ctx, "avatar-1").Return(avatar, nil)
		clock.EXPECT().Now().Return(now)
		idGen.EXPECT().NewID().Return("outbox-1", nil)
		transactor.EXPECT().RunInTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
		)
		tracer.EXPECT().ContextToMap(ctx).Return(map[string]string{"traceparent": "00-test"})
		writer.EXPECT().SoftDelete(ctx, "avatar-1", "user-1", now).Return(nil)
		outbox.EXPECT().CreateDeleted(ctx, dto.OutboxDeletedCreate{
			ID:           "outbox-1",
			CreatedAt:    now,
			TraceCarrier: map[string]string{"traceparent": "00-test"},
			Event: dto.AvatarDeletedEvent{
				AvatarID: "avatar-1",
				S3Keys:   []string{"user-1/avatar-1/original", "user-1/avatar-1/100x100/jpeg"},
			},
		}).Return(nil)

		uc := NewDeleteAvatar(reader, writer, outbox, transactor, idGen, clock, tracer)
		_, err := uc.Execute(ctx, dto.DeleteAvatarInput{AvatarID: "avatar-1", RequestUserID: "user-1"})
		require.NoError(t, err)
	})

	t.Run("forbidden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		reader := portmocks.NewMockAvatarReader(ctrl)
		tracer := pkgportmocks.NewMockTracer(ctrl)
		span := pkgportmocks.NewMockSpan(ctrl)
		span.EXPECT().Fail(gomock.Any())
		span.EXPECT().End()
		tracer.EXPECT().Start(ctx, gomock.Any()).Return(ctx, span)

		reader.EXPECT().FindByID(ctx, "avatar-1").Return(avatar, nil)

		uc := NewDeleteAvatar(
			reader,
			portmocks.NewMockAvatarWriter(ctrl),
			portmocks.NewMockOutboxWriter(ctrl),
			portmocks.NewMockTransactor(ctrl),
			portmocks.NewMockIDGenerator(ctrl),
			portmocks.NewMockClock(ctrl),
			tracer,
		)
		_, err := uc.Execute(ctx, dto.DeleteAvatarInput{AvatarID: "avatar-1", RequestUserID: "other-user"})
		assert.ErrorIs(t, err, application.ErrForbidden)
	})
}
