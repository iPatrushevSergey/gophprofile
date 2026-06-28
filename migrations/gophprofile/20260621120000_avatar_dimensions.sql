-- +goose Up
ALTER TABLE avatars
    ADD COLUMN IF NOT EXISTS width INTEGER,
    ADD COLUMN IF NOT EXISTS height INTEGER;

-- +goose Down
ALTER TABLE avatars
    DROP COLUMN IF EXISTS width,
    DROP COLUMN IF EXISTS height;
