package usecase

import (
	"context"

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
	clock        appport.Clock
}

// NewDeleteAvatar returns the delete avatar use case.
func NewDeleteAvatar(
	avatarReader appport.AvatarReader,
	avatarWriter appport.AvatarWriter,
	outboxWriter appport.OutboxWriter,
	transactor appport.Transactor,
	clock appport.Clock,
) appport.UseCase[dto.DeleteAvatarInput, struct{}] {
	return &DeleteAvatar{
		avatarReader: avatarReader,
		avatarWriter: avatarWriter,
		outboxWriter: outboxWriter,
		transactor:   transactor,
		clock:        clock,
	}
}

// Execute deletes an avatar.
func (uc *DeleteAvatar) Execute(ctx context.Context, in dto.DeleteAvatarInput) (struct{}, error) {
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

		return uc.outboxWriter.CreateDeleted(txCtx, dto.AvatarDeletedEvent{
			AvatarID: avatar.ID,
			S3Keys:   avatar.AllS3Keys(),
		})
	})
	if err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
