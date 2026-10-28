package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	m "backend/api_server/model"

	"github.com/google/uuid"
)

var ErrProfileNotFound = errors.New("profile not found")

func (r *PartRepo) GetProfile(ctx context.Context, userID uuid.UUID) (m.Profile, error) {
	var p m.Profile
	err := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.email, COALESCE(u.avatar, ''), COALESCE(p.bio, ''), u.created_at
		FROM users u
		LEFT JOIN profiles p ON p.user_id = u.id
		WHERE u.id = $1`, userID,
	).Scan(&p.ID, &p.Username, &p.Email, &p.Avatar, &p.Bio, &p.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.Profile{}, ErrProfileNotFound
		}
		return m.Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}
