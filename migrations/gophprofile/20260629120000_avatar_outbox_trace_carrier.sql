-- +goose Up
ALTER TABLE avatar_outbox
    ADD COLUMN IF NOT EXISTS trace_carrier JSONB;

-- +goose Down
ALTER TABLE avatar_outbox
    DROP COLUMN IF EXISTS trace_carrier;
