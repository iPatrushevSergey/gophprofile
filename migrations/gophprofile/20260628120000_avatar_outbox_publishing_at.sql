-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE avatar_outbox ADD COLUMN IF NOT EXISTS publishing_at TIMESTAMPTZ;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_avatar_outbox_publishing
    ON avatar_outbox (publishing_at)
    WHERE status = 'publishing';

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_avatar_outbox_publishing;

ALTER TABLE avatar_outbox DROP COLUMN IF EXISTS publishing_at;
