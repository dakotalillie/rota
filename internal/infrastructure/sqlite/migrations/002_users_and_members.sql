-- +goose Up
CREATE TABLE users (
    id    TEXT NOT NULL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    data  TEXT NOT NULL
);

CREATE TABLE members (
    id          TEXT NOT NULL PRIMARY KEY,
    rotation_id TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    data        TEXT NOT NULL,
    UNIQUE (rotation_id, user_id)
);

-- +goose Down
DROP TABLE members;
DROP TABLE users;
