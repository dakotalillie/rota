package sqlite

import (
	"context"
	"database/sql"
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

	var user domain.User
	err := db.QueryRowContext(ctx, `SELECT id, email, name FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Email, &user.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
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

	newUserID := newID("usr")
	_, err := db.ExecContext(ctx,
		`INSERT OR IGNORE INTO users (id, email, name) VALUES (?, ?, ?)`,
		newUserID, email, name,
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
