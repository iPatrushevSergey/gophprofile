package usecase

import (
	"context"

	pkgport "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/port"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	appport "github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/port"
)

// PurgeDeletedAvatar removes avatar files from object storage.
type PurgeDeletedAvatar struct {
	avatarStorage appport.AvatarStorage
	tracer        pkgport.Tracer
}

// NewPurgeDeletedAvatar returns the purge deleted avatar use case.
func NewPurgeDeletedAvatar(avatarStorage appport.AvatarStorage, tracer pkgport.Tracer) appport.UseCase[dto.PurgeDeletedAvatarInput, struct{}] {
	return &PurgeDeletedAvatar{avatarStorage: avatarStorage, tracer: tracer}
}

// Execute purges deleted avatar files.
func (uc *PurgeDeletedAvatar) Execute(ctx context.Context, in dto.PurgeDeletedAvatarInput) (_ struct{}, err error) {
	if in.AvatarID == "" || len(in.S3Keys) == 0 {
		return struct{}{}, application.ErrBadInput
	}

	ctx, span := uc.tracer.Start(ctx, pkgport.SpanConfig{
		Key:  "processing.application.purge_deleted_avatar.execute",
		Name: "purge deleted avatar",
		Kind: pkgport.SpanKindInternal,
		Attributes: []pkgport.Attribute{
			{Key: "avatar_id", Value: in.AvatarID},
			{Key: "s3_keys_count", Value: len(in.S3Keys)},
		},
	})
	defer func() {
		span.Fail(err)
		span.End()
	}()

	for _, key := range in.S3Keys {
		if err = uc.avatarStorage.Delete(ctx, key); err != nil {
			return struct{}{}, err
		}
	}

	return struct{}{}, nil
}
