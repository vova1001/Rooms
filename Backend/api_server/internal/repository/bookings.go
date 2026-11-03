package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	m "backend/api_server/model"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrEventStarted  = errors.New("event already started")
	ErrEventFull     = errors.New("event is full")
	ErrAlreadyBooked = errors.New("already booked")
)

func (r *PartRepo) JoinEvent(ctx context.Context, eventID, userID uuid.UUID) (m.Booking, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return m.Booking{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// FOR UPDATE блокирует строку события до конца транзакции:
	// все записи и одобрения на это событие идут строго по очереди
	var title string
	var startsAt time.Time
	var capacity int
	var ownerID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT title, starts_at, capacity, creator_id
		FROM events
		WHERE id = $1 AND status = 'active'
		FOR UPDATE`, eventID,
	).Scan(&title, &startsAt, &capacity, &ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.Booking{}, ErrEventNotFound
		}
		return m.Booking{}, fmt.Errorf("lock event: %w", err)
	}

	if !startsAt.After(time.Now()) {
		return m.Booking{}, ErrEventStarted
	}

	if capacity > 0 {
		var confirmed int
		err = tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM bookings
			WHERE event_id = $1 AND status = 'confirmed'`, eventID,
		).Scan(&confirmed)
		if err != nil {
			return m.Booking{}, fmt.Errorf("count bookings: %w", err)
		}
		if confirmed >= capacity {
			return m.Booking{}, ErrEventFull
		}
	}

	var b m.Booking
	err = tx.QueryRowContext(ctx, `
		INSERT INTO bookings (event_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (event_id, user_id) DO NOTHING
		RETURNING id, event_id, user_id, status, created_at, updated_at`, eventID, userID,
	).Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.Booking{}, ErrAlreadyBooked
		}
		return m.Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	// заявка улетает владельцу события
	_, err = tx.ExecContext(ctx, `
		INSERT INTO notifications (user_id, booking_id, type, text)
		VALUES ($1, $2, 'booking_request', $3)`,
		ownerID, b.ID, "Новая заявка на событие «"+title+"»",
	)
	if err != nil {
		return m.Booking{}, fmt.Errorf("insert notification: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return m.Booking{}, fmt.Errorf("commit: %w", err)
	}
	return b, nil
}
