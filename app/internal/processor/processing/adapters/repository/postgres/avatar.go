package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	postgreskit "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/postgres/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/postgres/converter/generated"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/application/dto"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/processor/processing/domain/vo"
)

// AvatarRepository is a PostgreSQL implementation of avatar repository.
type AvatarRepository struct {
	transactor *postgreskit.Transactor
	conv       converter.AvatarConverter
}

// NewAvatarRepository creates an avatar repository.
func NewAvatarRepository(transactor *postgreskit.Transactor) *AvatarRepository {
	return &AvatarRepository{
		transactor: transactor,
		conv:       &generated.AvatarConverterImpl{},
	}
}

// FindByID returns avatar by id.
func (r *AvatarRepository) FindByID(ctx context.Context, id string) (*entity.Avatar, error) {
	avatarID, err := uuid.Parse(id)
	if err != nil {
		return nil, application.ErrBadInput
	}

	var avatar entity.Avatar

	err = r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT
				id, user_id, file_name, mime_type, size_bytes, width, height, s3_key, thumbnail_s3_keys,
				upload_status, processing_status, created_at, updated_at, deleted_at
			FROM avatars
			WHERE id = $1 AND deleted_at IS NULL AND upload_status = $2`,
			avatarID, string(vo.UploadStatusCompleted),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRow, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[model.Avatar])
		if err != nil {
			return err
		}

		avatar, err = r.conv.AvatarModelToAvatarEntity(dbRow)
		return err
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}
	return &avatar, nil
}

// StartProcessing marks a pending avatar as processing.
func (r *AvatarRepository) StartProcessing(ctx context.Context, id string, updatedAt time.Time) error {
	avatarID, err := uuid.Parse(id)
	if err != nil {
		return application.ErrBadInput
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatars
			SET processing_status = $2, updated_at = $3
			WHERE id = $1
				AND deleted_at IS NULL
				AND upload_status = $4
				AND processing_status = $5`,
			avatarID, string(vo.ProcessingStatusProcessing), updatedAt,
			string(vo.UploadStatusCompleted), string(vo.ProcessingStatusPending),
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return application.ErrAlreadyProcessed
		}
		return nil
	})
}

// UpdateProcessingStatus updates processing status.
func (r *AvatarRepository) UpdateProcessingStatus(
	ctx context.Context,
	id string,
	status vo.ProcessingStatus,
	updatedAt time.Time,
) error {
	avatarID, err := uuid.Parse(id)
	if err != nil {
		return application.ErrBadInput
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatars
			SET processing_status = $2, updated_at = $3
			WHERE id = $1 AND deleted_at IS NULL AND upload_status = $4`,
			avatarID, string(status), updatedAt, string(vo.UploadStatusCompleted),
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return application.ErrNotFound
		}
		return nil
	})
}

// CompleteProcessing stores thumbnail keys and marks processing completed.
func (r *AvatarRepository) CompleteProcessing(ctx context.Context, in dto.CompleteProcessingInput) error {
	avatarID, err := uuid.Parse(in.AvatarID)
	if err != nil {
		return application.ErrBadInput
	}

	var thumbnailKeys json.RawMessage
	if len(in.ThumbnailS3Keys) > 0 {
		raw, err := json.Marshal(in.ThumbnailS3Keys)
		if err != nil {
			return err
		}
		thumbnailKeys = raw
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatars
			SET thumbnail_s3_keys = $2, width = $3, height = $4, processing_status = $5, updated_at = $6
			WHERE id = $1 AND deleted_at IS NULL AND upload_status = $7`,
			avatarID,
			thumbnailKeys,
			in.Width,
			in.Height,
			string(vo.ProcessingStatusCompleted),
			in.UpdatedAt,
			string(vo.UploadStatusCompleted),
		)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return application.ErrNotFound
		}
		return nil
	})
}
