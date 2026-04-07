-- +goose Up
CREATE INDEX idx_overrides_member ON overrides(member_id);

-- +goose Down
DROP INDEX idx_overrides_member;
