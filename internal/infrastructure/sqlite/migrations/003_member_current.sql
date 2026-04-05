-- +goose Up
ALTER TABLE members ADD COLUMN is_current INTEGER NOT NULL DEFAULT 0;
ALTER TABLE members ADD COLUMN became_current_at TEXT;

-- +goose Down
ALTER TABLE members DROP COLUMN is_current;
ALTER TABLE members DROP COLUMN became_current_at;
