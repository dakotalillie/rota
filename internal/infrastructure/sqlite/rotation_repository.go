package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/dakotalillie/rota/internal/domain"
)

type RotationRepository struct {
	db *sql.DB
}

func NewRotationRepository(db *sql.DB) *RotationRepository {
	return &RotationRepository{db: db}
}

func (r *RotationRepository) GetByID(ctx context.Context, id string) (*domain.Rotation, error) {
	row := dbFromContext(ctx, r.db).QueryRowContext(ctx, `SELECT id, data FROM rotations WHERE id = ?`, id)

	var rotID, rawData string
	if err := row.Scan(&rotID, &rawData); errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrRotationNotFound
	} else if err != nil {
		return nil, err
	}

	var rec rotationData
	if err := json.Unmarshal([]byte(rawData), &rec); err != nil {
		return nil, err
	}

	rot := &domain.Rotation{ID: rotID, Name: rec.Name}
	if rec.Cadence.Weekly != nil {
		rot.Cadence.Weekly = &domain.RotationCadenceWeekly{
			Day:      rec.Cadence.Weekly.Day,
			Time:     rec.Cadence.Weekly.Time,
			TimeZone: rec.Cadence.Weekly.TimeZone,
		}
	}
	return rot, nil
}

func (r *RotationRepository) UpsertRotation(ctx context.Context, rot *domain.Rotation) error {
	rec := rotationData{Name: rot.Name}
	if rot.Cadence.Weekly != nil {
		rec.Cadence.Weekly = &rotationCadenceWeekly{
			Day:      rot.Cadence.Weekly.Day,
			Time:     rot.Cadence.Weekly.Time,
			TimeZone: rot.Cadence.Weekly.TimeZone,
		}
	}

	blob, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	_, err = dbFromContext(ctx, r.db).ExecContext(
		ctx,
		`INSERT INTO rotations (id, data) VALUES (?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = excluded.data`,
		rot.ID, string(blob),
	)
	return err
}
