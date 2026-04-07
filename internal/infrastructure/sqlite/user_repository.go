package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/dakotalillie/rota/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	db := dbFromContext(ctx, r.db)

	var (
		user    domain.User
		rawData string
	)
	err := db.QueryRowContext(ctx, `SELECT id, email, data FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Email, &rawData)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	var rec userData
	if err := json.Unmarshal([]byte(rawData), &rec); err != nil {
		return nil, err
	}
	user.Name = rec.Name
	return &user, nil
}

func (r *UserRepository) CountMemberships(ctx context.Context, userID string) (int, error) {
	var count int
	err := dbFromContext(ctx, r.db).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM members WHERE user_id = ?`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	_, err := dbFromContext(ctx, r.db).ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

func (r *UserRepository) Create(ctx context.Context, name, email string) (*domain.User, error) {
	db := dbFromContext(ctx, r.db)

	rec := userData{Name: name}
	blob, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}

	newUserID := newID("usr")
	_, err = db.ExecContext(ctx,
		`INSERT OR IGNORE INTO users (id, email, data) VALUES (?, ?, ?)`,
		newUserID, email, string(blob),
	)
	if err != nil {
		return nil, err
	}

	var user domain.User
	user.Name = name
	user.Email = email
	if err := db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ?`, email).Scan(&user.ID); err != nil {
		return nil, err
	}
	return &user, nil
}
