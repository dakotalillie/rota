-- +goose Up
CREATE TABLE overrides (
    id          TEXT NOT NULL PRIMARY KEY,
    rotation_id TEXT NOT NULL REFERENCES rotations(id),
    member_id   TEXT NOT NULL REFERENCES members(id),
    start_time  TEXT NOT NULL,
    end_time    TEXT NOT NULL
);
CREATE INDEX idx_overrides_rotation_start ON overrides(rotation_id, start_time);
CREATE INDEX idx_overrides_rotation_end   ON overrides(rotation_id, end_time);

-- +goose Down
DROP TABLE overrides;
