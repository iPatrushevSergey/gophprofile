package usecase

import (
	"context"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// UploadAvatar uploads an avatar to the storage and database.
type UploadAvatar struct {
	avatarWriter  appport.AvatarWriter
	avatarStorage appport.AvatarStorage
	outboxWriter  appport.OutboxWriter
	transactor    appport.Transactor
	idGenerator   appport.IDGenerator
	clock         appport.Clock
	tracer        pkgport.Tracer
}

// NewUploadAvatar returns the upload avatar use case.
func NewUploadAvatar(
	avatarWriter appport.AvatarWriter,
	avatarStorage appport.AvatarStorage,
	outboxWriter appport.OutboxWriter,
	transactor appport.Transactor,
	idGenerator appport.IDGenerator,
	clock appport.Clock,
	tracer pkgport.Tracer,
) appport.UseCase[dto.UploadAvatarInput, dto.UploadAvatarOutput] {
	return &UploadAvatar{
		avatarWriter:  avatarWriter,
		avatarStorage: avatarStorage,
		outboxWriter:  outboxWriter,
		transactor:    transactor,
		idGenerator:   idGenerator,
		clock:         clock,
		tracer:        tracer,
	}
}

// Execute uploads an avatar.
func (uc *UploadAvatar) Execute(ctx context.Context, in dto.UploadAvatarInput) (out dto.UploadAvatarOutput, err error) {
	switch in.MimeType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return dto.UploadAvatarOutput{}, application.ErrBadInput
	}

	ctx, span := uc.tracer.Start(ctx, pkgport.SpanConfig{
		Key:  "avatar.application.upload_avatar.execute",
		Name: "upload avatar",
		Kind: pkgport.SpanKindInternal,
		Attributes: []pkgport.Attribute{
			{Key: "user_id", Value: in.UserID},
			{Key: "mime_type", Value: in.MimeType},
			{Key: "file_size", Value: len(in.Content)},
		},
	})
	defer func() {
		span.Fail(err)
		span.End()
	}()

	id, err := uc.idGenerator.NewID()
	if err != nil {
		return dto.UploadAvatarOutput{}, err
	}
	span.AddAttributes(pkgport.Attribute{Key: "avatar_id", Value: id})

	now := uc.clock.Now()
	s3Key := entity.OriginalObjectKey(in.UserID, id)
	avatar := entity.NewAvatar(
		id,
		in.UserID,
		in.FileName,
		in.MimeType,
		int64(len(in.Content)),
		s3Key,
		vo.UploadStatusUploading,
		now,
	)

	if err = uc.avatarWriter.Create(ctx, avatar); err != nil {
		return dto.UploadAvatarOutput{}, err
	}

	if err = uc.avatarStorage.Put(ctx, s3Key, in.Content, in.MimeType); err != nil {
		return dto.UploadAvatarOutput{}, err
	}

	err = uc.transactor.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.avatarWriter.MarkUploadCompleted(txCtx, avatar.ID, now); err != nil {
			return err
		}

		outboxID, err := uc.idGenerator.NewID()
		if err != nil {
			return err
		}

		return uc.outboxWriter.CreateUploaded(txCtx, dto.OutboxUploadedCreate{
			ID:           outboxID,
			CreatedAt:    now,
			TraceCarrier: uc.tracer.ContextToMap(txCtx),
			Event: dto.AvatarUploadedEvent{
				AvatarID: avatar.ID,
				UserID:   avatar.UserID,
				S3Key:    avatar.S3Key,
			},
		})
	})
	if err != nil {
		return dto.UploadAvatarOutput{}, err
	}

	avatar.UploadStatus = vo.UploadStatusCompleted

	return dto.UploadAvatarOutput{
		ID:               avatar.ID,
		UserID:           avatar.UserID,
		UploadStatus:     avatar.UploadStatus,
		ProcessingStatus: avatar.ProcessingStatus,
		CreatedAt:        avatar.CreatedAt,
	}, nil
}
