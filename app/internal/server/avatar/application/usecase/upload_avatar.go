package usecase

import (
	"context"

	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
)

// UploadAvatar uploads an avatar to the storage and database.
type UploadAvatar struct {
	avatarWriter  appport.AvatarWriter
	avatarStorage appport.AvatarStorage
	outboxWriter  appport.OutboxWriter
	transactor    appport.Transactor
	idGenerator   appport.IDGenerator
	clock         appport.Clock
}

// NewUploadAvatar returns the upload avatar use case.
func NewUploadAvatar(
	avatarWriter appport.AvatarWriter,
	avatarStorage appport.AvatarStorage,
	outboxWriter appport.OutboxWriter,
	transactor appport.Transactor,
	idGenerator appport.IDGenerator,
	clock appport.Clock,
) appport.UseCase[dto.UploadAvatarInput, dto.UploadAvatarOutput] {
	return &UploadAvatar{
		avatarWriter:  avatarWriter,
		avatarStorage: avatarStorage,
		outboxWriter:  outboxWriter,
		transactor:    transactor,
		idGenerator:   idGenerator,
		clock:         clock,
	}
}

// Execute uploads an avatar.
func (uc *UploadAvatar) Execute(ctx context.Context, in dto.UploadAvatarInput) (dto.UploadAvatarOutput, error) {
	switch in.MimeType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return dto.UploadAvatarOutput{}, application.ErrBadInput
	}

	id, err := uc.idGenerator.NewID()
	if err != nil {
		return dto.UploadAvatarOutput{}, err
	}

	s3Key := entity.OriginalObjectKey(in.UserID, id)
	if err := uc.avatarStorage.Put(ctx, s3Key, in.Content, in.MimeType); err != nil {
		return dto.UploadAvatarOutput{}, err
	}

	now := uc.clock.Now()
	avatar := entity.NewAvatar(id, in.UserID, in.FileName, in.MimeType, int64(len(in.Content)), s3Key, now)

	err = uc.transactor.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.avatarWriter.Create(txCtx, avatar); err != nil {
			return err
		}

		return uc.outboxWriter.EnqueueAvatarUploaded(txCtx, dto.AvatarUploadedEvent{
			AvatarID: avatar.ID,
			UserID:   avatar.UserID,
			S3Key:    avatar.S3Key,
		})
	})
	if err != nil {
		return dto.UploadAvatarOutput{}, err
	}

	return dto.UploadAvatarOutput{
		ID:               avatar.ID,
		UserID:           avatar.UserID,
		UploadStatus:     avatar.UploadStatus,
		ProcessingStatus: avatar.ProcessingStatus,
		CreatedAt:        avatar.CreatedAt,
	}, nil
}
