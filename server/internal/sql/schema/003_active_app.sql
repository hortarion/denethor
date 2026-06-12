-- +goose Up
ALTER TABLE users ADD COLUMN active_app TEXT DEFAULT 'console';

-- +goose Down
ALTER TABLE users DROP COLUMN active_app;
