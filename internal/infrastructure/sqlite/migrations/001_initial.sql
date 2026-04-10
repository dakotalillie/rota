-- +goose Up
CREATE TABLE rotations (
    id              TEXT NOT NULL PRIMARY KEY,
    name            TEXT NOT NULL,
    weekly_day      TEXT,
    weekly_time     TEXT,
    weekly_timezone TEXT
);

CREATE TABLE users (
    id    TEXT NOT NULL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name  TEXT NOT NULL
);

CREATE TABLE members (
    id                TEXT NOT NULL PRIMARY KEY,
    rotation_id       TEXT NOT NULL REFERENCES rotations(id),
    user_id           TEXT NOT NULL REFERENCES users(id),
    position          INTEGER NOT NULL,
    color             TEXT NOT NULL,
    is_current        INTEGER NOT NULL DEFAULT 0,
    became_current_at TEXT,
    UNIQUE (rotation_id, user_id)
);

CREATE TABLE overrides (
    id          TEXT NOT NULL PRIMARY KEY,
    rotation_id TEXT NOT NULL REFERENCES rotations(id),
    member_id   TEXT NOT NULL REFERENCES members(id),
    start_time  TEXT NOT NULL,
    end_time    TEXT NOT NULL
);
CREATE INDEX idx_overrides_rotation_start ON overrides(rotation_id, start_time);
CREATE INDEX idx_overrides_rotation_end   ON overrides(rotation_id, end_time);
CREATE INDEX idx_overrides_member         ON overrides(member_id);

-- +goose Down
DROP TABLE overrides;
DROP TABLE members;
DROP TABLE users;
DROP TABLE rotations;
