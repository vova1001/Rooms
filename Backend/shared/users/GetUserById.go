package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type RepositoryUser struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *RepositoryUser {
	return &RepositoryUser{
		db: db,
	}
}

func (r *RepositoryUser) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, username, created_at, avatar, email
		FROM users
		WHERE id = $1
	`

	var u User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.CreatedAt,
		&u.Avatar,
		&u.Email,
	)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("GetUserByID cancelled: %w", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %s", id)
		}

		return nil, fmt.Errorf("failed to get user by id %s: %w", id, err)
	}

	return &u, nil
}
