package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (r *MemberRepository) Create(ctx context.Context, rotationID, userID string, order int) (*domain.Member, error) {
	db := dbFromContext(ctx, r.db)

	memberID := newID("mem")
	rec := memberData{Order: order}
	blob, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx,
		`INSERT INTO members (id, rotation_id, user_id, data) VALUES (?, ?, ?, ?)`,
		memberID, rotationID, userID, string(blob),
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
		Order:      order,
	}, nil
}

func (r *MemberRepository) SetCurrentMember(ctx context.Context, memberID string, at time.Time) error {
	db := dbFromContext(ctx, r.db)
	_, err := db.ExecContext(ctx,
		`UPDATE members SET is_current = 1, became_current_at = ? WHERE id = ?`,
		at.UTC().Format(time.RFC3339), memberID,
	)
	return err
}
