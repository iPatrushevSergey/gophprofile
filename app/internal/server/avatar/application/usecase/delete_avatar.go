package usecase

import (
	"context"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
)

// DeleteAvatar soft-deletes avatar and enqueues cleanup event.
type DeleteAvatar struct {
	avatarReader appport.AvatarReader
	avatarWriter appport.AvatarWriter
	outboxWriter appport.OutboxWriter
	transactor   appport.Transactor
	idGenerator  appport.IDGenerator
	clock        appport.Clock
	tracer       pkgport.Tracer
	metrics      pkgport.Metrics
}

// NewDeleteAvatar returns the delete avatar use case.
func NewDeleteAvatar(
	avatarReader appport.AvatarReader,
	avatarWriter appport.AvatarWriter,
	outboxWriter appport.OutboxWriter,
	transactor appport.Transactor,
	idGenerator appport.IDGenerator,
	clock appport.Clock,
	tracer pkgport.Tracer,
	metrics pkgport.Metrics,
) appport.UseCase[dto.DeleteAvatarInput, struct{}] {
	return &DeleteAvatar{
		avatarReader: avatarReader,
		avatarWriter: avatarWriter,
		outboxWriter: outboxWriter,
		transactor:   transactor,
		idGenerator:  idGenerator,
		clock:        clock,
		tracer:       tracer,
		metrics:      metrics,
	}
}

// Execute deletes an avatar.
func (uc *DeleteAvatar) Execute(ctx context.Context, in dto.DeleteAvatarInput) (_ struct{}, err error) {
	defer func() {
		uc.metrics.RecordDelete(ctx, pkgport.MetricStatus(err, application.ErrBadInput))
	}()

	ctx, span := uc.tracer.Start(ctx, pkgport.SpanConfig{
		Key:  "avatar.application.delete_avatar.execute",
		Name: "delete avatar",
		Kind: pkgport.SpanKindInternal,
		Attributes: []pkgport.Attribute{
			{Key: "avatar_id", Value: in.AvatarID},
			{Key: "user_id", Value: in.RequestUserID},
		},
	})
	defer func() {
		span.Fail(err)
		span.End()
	}()

	avatar, err := uc.avatarReader.FindByID(ctx, in.AvatarID)
	if err != nil {
		return struct{}{}, err
	}

	if avatar.UserID != in.RequestUserID {
		return struct{}{}, application.ErrForbidden
	}

	now := uc.clock.Now()

	err = uc.transactor.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.avatarWriter.SoftDelete(txCtx, in.AvatarID, in.RequestUserID, now); err != nil {
			return err
		}

		outboxID, err := uc.idGenerator.NewID()
		if err != nil {
			return err
		}

		return uc.outboxWriter.CreateDeleted(txCtx, dto.OutboxDeletedCreate{
			ID:           outboxID,
			CreatedAt:    now,
			TraceCarrier: uc.tracer.ContextToMap(txCtx),
			Event: dto.AvatarDeletedEvent{
				AvatarID: avatar.ID,
				S3Keys:   avatar.AllS3Keys(),
			},
		})
	})
	if err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
