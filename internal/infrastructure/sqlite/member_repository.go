package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) CountByRotationID(ctx context.Context, rotationID string) (int, error) {
	var count int
	err := dbFromContext(ctx, r.db).QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM members WHERE rotation_id = ?`,
		rotationID,
	).Scan(&count)
	return count, err
}

func (r *MemberRepository) Create(ctx context.Context, rotationID, userID string, position int, color string) (*domain.Member, error) {
	db := dbFromContext(ctx, r.db)

	memberID := newID("mem")
	_, err := db.ExecContext(ctx,
		`INSERT INTO members (id, rotation_id, user_id, position, color) VALUES (?, ?, ?, ?, ?)`,
		memberID, rotationID, userID, position, color,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, domain.ErrMemberAlreadyExists
		}
		return nil, err
	}

	return &domain.Member{
		ID:         memberID,
		RotationID: rotationID,
		User:       domain.User{ID: userID},
		Position:   position,
		Color:      color,
	}, nil
}

func (r *MemberRepository) GetByID(ctx context.Context, rotationID, memberID string) (*domain.Member, error) {
	db := dbFromContext(ctx, r.db)

	var m domain.Member
	err := db.QueryRowContext(ctx,
		`SELECT id, rotation_id, user_id, position, color FROM members WHERE id = ? AND rotation_id = ?`,
		memberID, rotationID,
	).Scan(&m.ID, &m.RotationID, &m.User.ID, &m.Position, &m.Color)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrMemberNotFound
	} else if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MemberRepository) Delete(ctx context.Context, memberID string) error {
	_, err := dbFromContext(ctx, r.db).ExecContext(ctx, `DELETE FROM members WHERE id = ?`, memberID)
	return err
}

func (r *MemberRepository) SetCurrentMember(ctx context.Context, rotationID string, memberID string, at time.Time) error {
	db := dbFromContext(ctx, r.db)
	_, err := db.ExecContext(ctx, `
		UPDATE members
		SET is_current        = CASE WHEN id = ? THEN 1 ELSE 0 END,
		    became_current_at = CASE WHEN id = ? THEN ? ELSE became_current_at END
		WHERE rotation_id = ?
	`, memberID, memberID, at.UTC().Format(time.RFC3339), rotationID)
	return err
}

func (r *MemberRepository) ReorderMembers(ctx context.Context, rotationID string, memberIDs []string) error {
	db := dbFromContext(ctx, r.db)
	for i, memberID := range memberIDs {
		_, err := db.ExecContext(ctx,
			`UPDATE members SET position = ? WHERE id = ? AND rotation_id = ?`,
			i+1, memberID, rotationID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
