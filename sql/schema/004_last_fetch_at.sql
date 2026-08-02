-- +goose Up
ALTER TABLE feed
ADD COLUMN last_fetch_at TIMESTAMP;

-- +goose Down
ALTER TABLE feed
DROP COLUMN last_fetch_at;