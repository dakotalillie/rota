package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type OverrideRepository struct {
	db *sql.DB
}

func NewOverrideRepository(db *sql.DB) *OverrideRepository {
	return &OverrideRepository{db: db}
}

func (r *OverrideRepository) Create(ctx context.Context, rotationID, memberID string, start, end time.Time) (*domain.Override, error) {
	id := newID("ovr")
	_, err := dbFromContext(ctx, r.db).ExecContext(ctx,
		`INSERT INTO overrides (id, rotation_id, member_id, start_time, end_time) VALUES (?, ?, ?, ?, ?)`,
		id, rotationID, memberID,
		start.UTC().Format(time.RFC3339),
		end.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	return &domain.Override{
		ID:         id,
		RotationID: rotationID,
		Member:     domain.Member{ID: memberID},
		Start:      start.UTC(),
		End:        end.UTC(),
	}, nil
}

func (r *OverrideRepository) HasOverlapping(ctx context.Context, rotationID string, start, end time.Time) (bool, error) {
	var count int
	err := dbFromContext(ctx, r.db).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM overrides WHERE rotation_id = ? AND start_time < ? AND end_time > ?`,
		rotationID,
		end.UTC().Format(time.RFC3339),
		start.UTC().Format(time.RFC3339),
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
