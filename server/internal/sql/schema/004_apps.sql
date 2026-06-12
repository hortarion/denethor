-- +goose Up

CREATE TABLE apps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    app_name TEXT NOT NULL UNIQUE
);

-- +goose Down

DROP TABLE apps;
