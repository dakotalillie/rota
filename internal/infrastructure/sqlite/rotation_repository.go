package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/dakotalillie/rota/internal/domain"
)

type RotationRepository struct {
	db *sql.DB
}

func (r *RotationRepository) GetRotationByID(id string) (*domain.Rotation, error) {
	row := r.db.QueryRow(`SELECT id, data FROM rotations WHERE id = ?`, id)

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

	return &domain.Rotation{
		ID:   rotID,
		Name: rec.Name,
		Cadence: domain.RotationCadence{
			Weekly: &domain.RotationCadenceWeekly{
				Day:      rec.Cadence.Weekly.Day,
				Time:     rec.Cadence.Weekly.Time,
				TimeZone: rec.Cadence.Weekly.TimeZone,
			},
		},
	}, nil
}

func NewRotationRepository(db *sql.DB) *RotationRepository {
	return &RotationRepository{db: db}
}
