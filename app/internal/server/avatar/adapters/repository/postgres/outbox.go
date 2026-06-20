package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	postgreskit "github.com/iPatrushevSergey/gophprofile/app/internal/pkg/adapters/repository/postgres"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/converter/generated"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/adapters/repository/postgres/model"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/application/dto"
	"github.com/iPatrushevSergey/gophprofile/app/internal/server/avatar/domain/vo"
)

// OutboxRepository is a PostgreSQL implementation of the avatar outbox repository.
type OutboxRepository struct {
	transactor *postgreskit.Transactor
	conv       converter.OutboxConverter
}

// NewOutboxRepository creates an outbox repository.
func NewOutboxRepository(transactor *postgreskit.Transactor) *OutboxRepository {
	return &OutboxRepository{
		transactor: transactor,
		conv:       &generated.OutboxConverterImpl{},
	}
}

// CreateUploaded inserts a pending avatar uploaded outbox event.
func (r *OutboxRepository) CreateUploaded(ctx context.Context, event dto.OutboxUploadedCreate) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		dbRow, err := r.conv.OutboxUploadedCreateToOutboxEventModel(event)
		if err != nil {
			return err
		}

		_, err = q.Exec(ctx, `
			INSERT INTO avatar_outbox (id, event_type, payload, created_at)
			VALUES ($1, $2, $3, $4)`,
			dbRow.ID, string(vo.OutboxEventAvatarUploaded), dbRow.Payload, dbRow.CreatedAt,
		)
		return err
	})
}

// CreateDeleted inserts a pending avatar deleted outbox event.
func (r *OutboxRepository) CreateDeleted(ctx context.Context, event dto.OutboxDeletedCreate) error {
	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		dbRow, err := r.conv.OutboxDeletedCreateToOutboxEventModel(event)
		if err != nil {
			return err
		}

		_, err = q.Exec(ctx, `
			INSERT INTO avatar_outbox (id, event_type, payload, created_at)
			VALUES ($1, $2, $3, $4)`,
			dbRow.ID, string(vo.OutboxEventAvatarDeleted), dbRow.Payload, dbRow.CreatedAt,
		)
		return err
	})
}

// MarkPublished marks an outbox record as published.
func (r *OutboxRepository) MarkPublished(ctx context.Context, id string, publishedAt time.Time) error {
	outboxID, err := uuid.Parse(id)
	if err != nil {
		return application.ErrBadInput
	}

	return r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		tag, err := q.Exec(ctx, `
			UPDATE avatar_outbox
			SET status = $2, published_at = $3, attempts = attempts + 1
			WHERE id = $1`,
			outboxID, string(vo.OutboxStatusPublished), publishedAt,
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

// ListPending returns pending outbox events ordered by creation time.
func (r *OutboxRepository) ListPending(ctx context.Context, limit int) ([]dto.OutboxEvent, error) {
	var events []dto.OutboxEvent

	err := r.transactor.DoWithRetry(ctx, func() error {
		q := r.transactor.GetQuerier(ctx)

		rows, err := q.Query(ctx, `
			SELECT id, event_type, payload, status, created_at, published_at, attempts
			FROM avatar_outbox
			WHERE status = $1
			ORDER BY created_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED`,
			string(vo.OutboxStatusPending), limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.OutboxEvent])
		if err != nil {
			return err
		}

		events = make([]dto.OutboxEvent, 0, len(dbRows))
		for _, dbRow := range dbRows {
			event, err := converter.OutboxEventModelToOutboxEventDto(dbRow)
			if err != nil {
				return err
			}
			events = append(events, event)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return events, nil
}
