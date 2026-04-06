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
	row := dbFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT r.id, r.data, m.id, m.rotation_id, m.data, u.id, u.email, u.data
		FROM rotations r
		LEFT JOIN members m ON m.rotation_id = r.id AND m.is_current = 1
		LEFT JOIN users u ON u.id = m.user_id
		WHERE r.id = ?
	`, id)

	var (
		rotID, rawRotData          string
		memID, memRotID, rawMem    sql.NullString
		userID, userEmail, rawUser sql.NullString
	)
	if err := row.Scan(&rotID, &rawRotData, &memID, &memRotID, &rawMem, &userID, &userEmail, &rawUser); errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrRotationNotFound
	} else if err != nil {
		return nil, err
	}

	var rec rotationData
	if err := json.Unmarshal([]byte(rawRotData), &rec); err != nil {
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

	if memID.Valid {
		var mRec memberData
		if err := json.Unmarshal([]byte(rawMem.String), &mRec); err != nil {
			return nil, err
		}
		var uRec userData
		if err := json.Unmarshal([]byte(rawUser.String), &uRec); err != nil {
			return nil, err
		}
		rot.CurrentMember = &domain.Member{
			ID:         memID.String,
			RotationID: memRotID.String,
			Order:      mRec.Order,
			User: domain.User{
				ID:    userID.String,
				Email: userEmail.String,
				Name:  uRec.Name,
			},
		}
	}

	rows, err := dbFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT m.id, m.rotation_id, m.data, u.id, u.email, u.data
		FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE m.rotation_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var (
			mID, mRotID, rawM string
			uID, uEmail, rawU string
		)
		if err := rows.Scan(&mID, &mRotID, &rawM, &uID, &uEmail, &rawU); err != nil {
			return nil, err
		}
		var mRec memberData
		if err := json.Unmarshal([]byte(rawM), &mRec); err != nil {
			return nil, err
		}
		var uRec userData
		if err := json.Unmarshal([]byte(rawU), &uRec); err != nil {
			return nil, err
		}
		rot.Members = append(rot.Members, domain.Member{
			ID:         mID,
			RotationID: mRotID,
			Order:      mRec.Order,
			User: domain.User{
				ID:    uID,
				Email: uEmail,
				Name:  uRec.Name,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
