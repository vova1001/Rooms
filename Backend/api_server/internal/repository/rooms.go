package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (r *PartRepo) GetRoomLock(ctx context.Context, roomID uuid.UUID) (bool, uuid.UUID, sql.NullString, error) {
	var isPrivate bool
	var ownerID uuid.UUID
	var passHash sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT is_private, owner_id, password_hash
		FROM rooms
		WHERE id = $1`, roomID,
	).Scan(&isPrivate, &ownerID, &passHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, uuid.Nil, sql.NullString{}, ErrRoomNotFound
		}
		return false, uuid.Nil, sql.NullString{}, fmt.Errorf("get room lock: %w", err)
	}
	return isPrivate, ownerID, passHash, nil
}
