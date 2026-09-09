package message

import (
	m "backend/gateway_server/models"
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) SaveMessage(ctx context.Context, msg *m.Message) (*m.Message, error) {
	err := r.db.QueryRowContext(ctx,
		`
		INSERT INTO messages(
			room_id,
			user_id,
			msg
		)
		VALUES($1,$2,$3)
		RETURNING id`, msg.RoomID, msg.UserID, msg.Text).Scan(&msg.ID)

	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	return msg, nil
}
