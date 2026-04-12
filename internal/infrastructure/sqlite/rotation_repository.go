package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type RotationRepository struct {
	db *sql.DB
}

func NewRotationRepository(db *sql.DB) *RotationRepository {
	return &RotationRepository{db: db}
}

func (r *RotationRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := dbFromContext(ctx, r.db).QueryRowContext(ctx, `SELECT COUNT(*) FROM rotations`).Scan(&count)
	return count, err
}

func (r *RotationRepository) Create(ctx context.Context, rot *domain.Rotation) (*domain.Rotation, error) {
	rot.ID = newID("rot")

	var weeklyDay, weeklyTime, weeklyTimezone sql.NullString
	if rot.Cadence.Weekly != nil {
		weeklyDay = sql.NullString{String: rot.Cadence.Weekly.Day, Valid: true}
		weeklyTime = sql.NullString{String: rot.Cadence.Weekly.Time, Valid: true}
		weeklyTimezone = sql.NullString{String: rot.Cadence.Weekly.TimeZone, Valid: true}
	}

	_, err := dbFromContext(ctx, r.db).ExecContext(ctx,
		`INSERT INTO rotations (id, name, weekly_day, weekly_time, weekly_timezone) VALUES (?, ?, ?, ?, ?)`,
		rot.ID, rot.Name, weeklyDay, weeklyTime, weeklyTimezone,
	)
	if err != nil {
		return nil, err
	}
	return rot, nil
}

func (r *RotationRepository) Delete(ctx context.Context, id string) error {
	result, err := dbFromContext(ctx, r.db).ExecContext(ctx,
		`DELETE FROM rotations WHERE id = ?`,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrRotationNotFound
	}

	return nil
}

func (r *RotationRepository) GetByID(ctx context.Context, id string) (*domain.Rotation, error) {
	row := dbFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT r.id, r.name, r.weekly_day, r.weekly_time, r.weekly_timezone,
		       m.id, m.rotation_id, m.position, m.color, m.became_current_at,
		       u.id, u.email, u.name
		FROM rotations r
		LEFT JOIN members m ON m.rotation_id = r.id AND m.is_current = 1
		LEFT JOIN users u ON u.id = m.user_id
		WHERE r.id = ?
	`, id)

	var (
		rotID, rotName                        string
		weeklyDay, weeklyTime, weeklyTimezone sql.NullString
		memID, memRotID                       sql.NullString
		memPosition                           sql.NullInt64
		memColor, becameCurrentAt             sql.NullString
		userID, userEmail, userName           sql.NullString
	)
	if err := row.Scan(
		&rotID, &rotName, &weeklyDay, &weeklyTime, &weeklyTimezone,
		&memID, &memRotID, &memPosition, &memColor, &becameCurrentAt,
		&userID, &userEmail, &userName,
	); errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrRotationNotFound
	} else if err != nil {
		return nil, err
	}

	rot := &domain.Rotation{ID: rotID, Name: rotName}
	if weeklyDay.Valid && weeklyTime.Valid && weeklyTimezone.Valid {
		rot.Cadence.Weekly = &domain.RotationCadenceWeekly{
			Day:      weeklyDay.String,
			Time:     weeklyTime.String,
			TimeZone: weeklyTimezone.String,
		}
	}

	if memID.Valid {
		rot.ScheduledMember = &domain.Member{
			ID:         memID.String,
			RotationID: memRotID.String,
			Position:   int(memPosition.Int64),
			Color:      memColor.String,
			User: domain.User{
				ID:    userID.String,
				Email: userEmail.String,
				Name:  userName.String,
			},
		}
		if becameCurrentAt.Valid {
			t, err := time.Parse(time.RFC3339, becameCurrentAt.String)
			if err != nil {
				return nil, fmt.Errorf("parsing became_current_at: %w", err)
			}
			rot.ScheduledMember.BecameCurrentAt = t.UTC()
		}
	}

	rows, err := dbFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT m.id, m.rotation_id, m.position, m.color, u.id, u.email, u.name
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
			mID, mRotID        string
			mPosition          int
			mColor             string
			uID, uEmail, uName string
		)
		if err := rows.Scan(&mID, &mRotID, &mPosition, &mColor, &uID, &uEmail, &uName); err != nil {
			return nil, err
		}
		rot.Members = append(rot.Members, domain.Member{
			ID:         mID,
			RotationID: mRotID,
			Position:   mPosition,
			Color:      mColor,
			User: domain.User{
				ID:    uID,
				Email: uEmail,
				Name:  uName,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rot, nil
}

func (r *RotationRepository) List(ctx context.Context) ([]*domain.Rotation, error) {
	rows, err := dbFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT r.id, r.name, r.weekly_day, r.weekly_time, r.weekly_timezone,
		       m.id, m.rotation_id, m.position, m.color, m.became_current_at,
		       u.id, u.email, u.name
		FROM rotations r
		LEFT JOIN members m ON m.rotation_id = r.id AND m.is_current = 1
		LEFT JOIN users u ON u.id = m.user_id
		ORDER BY r.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var rotations []*domain.Rotation
	for rows.Next() {
		var (
			rotID, rotName                        string
			weeklyDay, weeklyTime, weeklyTimezone sql.NullString
			memID, memRotID                       sql.NullString
			memPosition                           sql.NullInt64
			memColor, becameCurrentAt             sql.NullString
			userID, userEmail, userName           sql.NullString
		)
		if err := rows.Scan(
			&rotID, &rotName, &weeklyDay, &weeklyTime, &weeklyTimezone,
			&memID, &memRotID, &memPosition, &memColor, &becameCurrentAt,
			&userID, &userEmail, &userName,
		); err != nil {
			return nil, err
		}

		rot := &domain.Rotation{ID: rotID, Name: rotName}
		if weeklyDay.Valid && weeklyTime.Valid && weeklyTimezone.Valid {
			rot.Cadence.Weekly = &domain.RotationCadenceWeekly{
				Day:      weeklyDay.String,
				Time:     weeklyTime.String,
				TimeZone: weeklyTimezone.String,
			}
		}

		if memID.Valid {
			rot.ScheduledMember = &domain.Member{
				ID:         memID.String,
				RotationID: memRotID.String,
				Position:   int(memPosition.Int64),
				Color:      memColor.String,
				User: domain.User{
					ID:    userID.String,
					Email: userEmail.String,
					Name:  userName.String,
				},
			}
			if becameCurrentAt.Valid {
				t, err := time.Parse(time.RFC3339, becameCurrentAt.String)
				if err != nil {
					return nil, fmt.Errorf("parsing became_current_at: %w", err)
				}
				rot.ScheduledMember.BecameCurrentAt = t.UTC()
			}
		}

		rotations = append(rotations, rot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if rotations == nil {
		rotations = []*domain.Rotation{}
	}
	return rotations, nil
}

func (r *RotationRepository) UpsertRotation(ctx context.Context, rot *domain.Rotation) error {
	var weeklyDay, weeklyTime, weeklyTimezone sql.NullString
	if rot.Cadence.Weekly != nil {
		weeklyDay = sql.NullString{String: rot.Cadence.Weekly.Day, Valid: true}
		weeklyTime = sql.NullString{String: rot.Cadence.Weekly.Time, Valid: true}
		weeklyTimezone = sql.NullString{String: rot.Cadence.Weekly.TimeZone, Valid: true}
	}

	_, err := dbFromContext(ctx, r.db).ExecContext(
		ctx,
		`INSERT INTO rotations (id, name, weekly_day, weekly_time, weekly_timezone) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   name = excluded.name,
		   weekly_day = excluded.weekly_day,
		   weekly_time = excluded.weekly_time,
		   weekly_timezone = excluded.weekly_timezone`,
		rot.ID, rot.Name, weeklyDay, weeklyTime, weeklyTimezone,
	)
	return err
}
