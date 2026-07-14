-- +goose NO TRANSACTION
-- +goose Up
CREATE TABLE IF NOT EXISTS avatar_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_avatar_outbox_pending
    ON avatar_outbox (status, created_at)
    WHERE status = 'pending';

-- +goose Down
DROP TABLE IF EXISTS avatar_outbox;
