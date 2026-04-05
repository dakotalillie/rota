-- +goose Up
CREATE TABLE rotations (
    id   TEXT NOT NULL PRIMARY KEY,
    data TEXT NOT NULL
);

-- +goose Down
DROP TABLE rotations;
