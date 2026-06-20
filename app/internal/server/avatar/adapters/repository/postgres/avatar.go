package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	postgreskit "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/generated"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/entity"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// AvatarRepository is a PostgreSQL implementation of the avatar repository.
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
				id, user_id, file_name, mime_type, size_bytes, s3_key, thumbnail_s3_keys,
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

// ListByUserID returns avatars for a user.
func (r *AvatarRepository) ListByUserID(ctx context.Context, userID string) ([]entity.Avatar, error) {
	var avatars []entity.Avatar

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT
				id, user_id, file_name, mime_type, size_bytes, s3_key, thumbnail_s3_keys,
				upload_status, processing_status, created_at, updated_at, deleted_at
			FROM avatars
			WHERE user_id = $1 AND deleted_at IS NULL AND upload_status = $2
			ORDER BY created_at DESC`,
			userID, string(vo.UploadStatusCompleted),
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Avatar])
		if err != nil {
			return err
		}

		avatars = make([]entity.Avatar, 0, len(dbRows))
		for _, dbRow := range dbRows {
			item, err := r.conv.AvatarModelToAvatarEntity(dbRow)
			if err != nil {
				return err
			}
			avatars = append(avatars, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return avatars, nil
}

// ListExpiredUploading returns uploading avatars created before the cutoff time.
func (r *AvatarRepository) ListExpiredUploading(ctx context.Context, before time.Time) ([]entity.Avatar, error) {
	var avatars []entity.Avatar

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT
				id, user_id, file_name, mime_type, size_bytes, s3_key, thumbnail_s3_keys,
				upload_status, processing_status, created_at, updated_at, deleted_at
			FROM avatars
			WHERE upload_status = $1 AND created_at < $2 AND deleted_at IS NULL`,
			string(vo.UploadStatusUploading), before,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.Avatar])
		if err != nil {
			return err
		}

		avatars = make([]entity.Avatar, 0, len(dbRows))
		for _, dbRow := range dbRows {
			item, err := r.conv.AvatarModelToAvatarEntity(dbRow)
			if err != nil {
				return err
			}
			avatars = append(avatars, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return avatars, nil
}

// Create inserts a new avatar.
func (r *AvatarRepository) Create(ctx context.Context, avatar *entity.Avatar) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		dbAvatar, err := r.conv.AvatarEntityToAvatarModel(*avatar)
		if err != nil {
			return err
		}

		_, err = q.Exec(ctx, `
			INSERT INTO avatars (
				id, user_id, file_name, mime_type, size_bytes, s3_key, thumbnail_s3_keys,
				upload_status, processing_status, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			dbAvatar.ID, dbAvatar.UserID, dbAvatar.FileName, dbAvatar.MimeType, dbAvatar.SizeBytes,
			dbAvatar.S3Key, dbAvatar.ThumbnailS3Keys, dbAvatar.UploadStatus, dbAvatar.ProcessingStatus,
			dbAvatar.CreatedAt, dbAvatar.UpdatedAt,
		)
		return err
	})
}

// MarkUploadCompleted marks avatar upload as completed.
func (r *AvatarRepository) MarkUploadCompleted(ctx context.Context, id string, updatedAt time.Time) error {
	avatarID, err := uuid.Parse(id)
	if err != nil {
		return application.ErrBadInput
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatars
			SET upload_status = $2, updated_at = $3
			WHERE id = $1 AND upload_status = $4 AND deleted_at IS NULL`,
			avatarID, string(vo.UploadStatusCompleted), updatedAt, string(vo.UploadStatusUploading),
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

// MarkUploadFailed marks avatar upload as failed.
func (r *AvatarRepository) MarkUploadFailed(ctx context.Context, id string, updatedAt time.Time) error {
	avatarID, err := uuid.Parse(id)
	if err != nil {
		return application.ErrBadInput
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatars
			SET upload_status = $2, updated_at = $3
			WHERE id = $1 AND upload_status = $4 AND deleted_at IS NULL`,
			avatarID, string(vo.UploadStatusFailed), updatedAt, string(vo.UploadStatusUploading),
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

// SoftDelete marks avatar as deleted.
func (r *AvatarRepository) SoftDelete(ctx context.Context, id, userID string, deletedAt time.Time) error {
	avatarID, err := uuid.Parse(id)
	if err != nil {
		return application.ErrBadInput
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatars
			SET deleted_at = $3, updated_at = $3
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
			avatarID, userID, deletedAt,
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

// Ping checks database connectivity.
func (r *AvatarRepository) Ping(ctx context.Context) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)
		return q.QueryRow(ctx, `SELECT 1`).Scan(new(int))
	})
}
