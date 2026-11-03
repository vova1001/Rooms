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

func (r *PartRepo) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]m.MyBooking, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT b.id, b.event_id, b.user_id, b.status, b.created_at, b.updated_at,
		       e.id, e.title, e.description, e.starts_at, e.ends_at,
		       e.price_minor, e.currency, e.capacity,
		       r.id, r.name, r.is_private,
		       (SELECT COUNT(*) FROM bookings c WHERE c.event_id = e.id AND c.status = 'confirmed')
		FROM bookings b
		JOIN events e ON e.id = b.event_id
		JOIN rooms r ON r.id = e.room_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]m.MyBooking, 0)
	for rows.Next() {
		var b m.MyBooking
		var isPrivate bool
		if err := rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.UpdatedAt,
			&b.Event.ID, &b.Event.Title, &b.Event.Description, &b.Event.StartsAt, &b.Event.EndsAt,
			&b.Event.PriceMinor, &b.Event.Currency, &b.Event.Capacity,
			&b.Event.Room.ID, &b.Event.Room.Name, &isPrivate,
			&b.Event.ParticipantsCount); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		b.Event.IsFree = b.Event.PriceMinor == 0
		b.Event.IsFull = b.Event.Capacity > 0 && b.Event.ParticipantsCount >= b.Event.Capacity
		b.Event.Room.Visibility = "open"
		if isPrivate {
			b.Event.Room.Visibility = "closed"
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}
