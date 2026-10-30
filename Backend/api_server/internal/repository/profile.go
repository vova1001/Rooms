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

func (r *PartRepo) ChangeProfile(ctx context.Context, userId uuid.UUID, req m.UpdateProfileReq) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx:%w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `
		UPDATE users SET 
			username = COALESCE($1, username),
			updated_at = NOW()
		WHERE id = $2`,
		req.Username, userId)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrProfileNotFound
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE profiles SET
			bio = COALESCE($1, bio),
			updated_at = NOW()
		WHERE user_id = $2`,
		req.Bio, userId)

	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("err commit: %w", err)
	}
	return nil
}
