package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func (r *OverrideRepository) Delete(ctx context.Context, rotationID, overrideID string) error {
	result, err := dbFromContext(ctx, r.db).ExecContext(ctx,
		`DELETE FROM overrides WHERE id = ? AND rotation_id = ?`,
		overrideID,
		rotationID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrOverrideNotFound
	}

	return nil
}

func (r *OverrideRepository) DeleteByMemberID(ctx context.Context, memberID string) error {
	_, err := dbFromContext(ctx, r.db).ExecContext(ctx,
		`DELETE FROM overrides WHERE member_id = ?`,
		memberID,
	)
	return err
}

func (r *OverrideRepository) ListByRotationID(ctx context.Context, rotationID string, now time.Time) ([]domain.Override, error) {
	rows, err := dbFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT o.id, o.start_time, o.end_time, m.id, m.rotation_id, m.data, u.id, u.email, u.data
		FROM overrides o
		JOIN members m ON o.member_id = m.id
		JOIN users u ON m.user_id = u.id
		WHERE o.rotation_id = ? AND o.end_time > ?
		ORDER BY o.start_time
	`, rotationID, now.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var overrides []domain.Override
	for rows.Next() {
		var (
			oID, oStart, oEnd string
			mID, mRotID, rawM string
			uID, uEmail, rawU string
		)
		if err := rows.Scan(&oID, &oStart, &oEnd, &mID, &mRotID, &rawM, &uID, &uEmail, &rawU); err != nil {
			return nil, err
		}
		start, err := time.Parse(time.RFC3339, oStart)
		if err != nil {
			return nil, fmt.Errorf("parsing override start_time: %w", err)
		}
		end, err := time.Parse(time.RFC3339, oEnd)
		if err != nil {
			return nil, fmt.Errorf("parsing override end_time: %w", err)
		}
		var mRec memberData
		if err := json.Unmarshal([]byte(rawM), &mRec); err != nil {
			return nil, err
		}
		var uRec userData
		if err := json.Unmarshal([]byte(rawU), &uRec); err != nil {
			return nil, err
		}
		overrides = append(overrides, domain.Override{
			ID:         oID,
			RotationID: rotationID,
			Start:      start.UTC(),
			End:        end.UTC(),
			Member: domain.Member{
				ID:         mID,
				RotationID: mRotID,
				Order:      mRec.Order,
				Color:      mRec.Color,
				User: domain.User{
					ID:    uID,
					Email: uEmail,
					Name:  uRec.Name,
				},
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if overrides == nil {
		overrides = []domain.Override{}
	}
	return overrides, nil
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
