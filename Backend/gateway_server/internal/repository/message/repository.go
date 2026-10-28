package message

import (
	m "backend/gateway_server/models"
	"context"
	"database/sql"
	"fmt"
	"time"
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
			msg,
			created_at
		)
		VALUES($1,$2,$3,$4)
		RETURNING id`, msg.RoomID, msg.UserID, msg.Text, msg.CreatedAt).Scan(&msg.ID)

	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	return msg, nil
}

func (r *Repository) GetMessages(ctx context.Context, roomID string, before *time.Time, limit int) ([]m.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.room_id, m.user_id, u.username, COALESCE(u.avatar, ''), m.msg, m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.room_id = $1
  			AND m.deleted_at IS NULL
  			AND ($2::timestamptz IS NULL OR m.created_at < $2)
		ORDER BY m.created_at DESC
		LIMIT $3`, roomID, before, limit)
	if err != nil {
		return nil, fmt.Errorf("select messages: %w", err)
	}
	defer rows.Close()

	messages := []m.Message{}
	for rows.Next() {
		var msg m.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.UserName, &msg.Avatar, &msg.Text, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return messages, nil
}
