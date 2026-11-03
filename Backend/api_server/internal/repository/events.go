package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	m "backend/api_server/model"
)

var ErrRoomNotFound = errors.New("room not found")

func (r *PartRepo) CreateEvent(ctx context.Context, req m.CreateEventReq) (m.Event, error) {
	var ev m.Event

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO events (room_id, creator_id, title, description, starts_at, ends_at, price_minor, capacity)
		SELECT id, $2, $3, $4, $5, $6, $7, $8
		FROM rooms
		WHERE id = $1 AND owner_id = $2
		RETURNING id, room_id, creator_id, title, description, starts_at, ends_at, price_minor, currency, capacity, status, created_at`,
		req.RoomID, req.CreatorID, req.Title, req.Description, req.StartsAt, req.EndsAt, req.PriceMinor, req.Capacity,
	).Scan(&ev.ID, &ev.RoomID, &ev.CreatorID, &ev.Title, &ev.Description,
		&ev.StartsAt, &ev.EndsAt, &ev.PriceMinor, &ev.Currency, &ev.Capacity, &ev.Status, &ev.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.Event{}, ErrRoomNotFound
		}
		return m.Event{}, fmt.Errorf("create event: %w", err)
	}

	ev.IsFree = ev.PriceMinor == 0
	return ev, nil
}

func (r *PartRepo) ListEvents(ctx context.Context, dayStart, dayEnd time.Time) ([]m.EventItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.id, e.title, e.description, e.starts_at, e.ends_at,
		       e.price_minor, e.currency, e.capacity,
		       r.id, r.name, r.is_private,
		       (SELECT COUNT(*) FROM bookings b WHERE b.event_id = e.id AND b.status = 'confirmed')
		FROM events e
		JOIN rooms r ON r.id = e.room_id
		WHERE e.starts_at < $2 AND e.ends_at > $1
		ORDER BY e.starts_at`,
		dayStart, dayEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]m.EventItem, 0)
	for rows.Next() {
		var ev m.EventItem
		var isPrivate bool
		if err := rows.Scan(&ev.ID, &ev.Title, &ev.Description, &ev.StartsAt, &ev.EndsAt,
			&ev.PriceMinor, &ev.Currency, &ev.Capacity,
			&ev.Room.ID, &ev.Room.Name, &isPrivate,
			&ev.ParticipantsCount); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		ev.IsFree = ev.PriceMinor == 0
		ev.IsFull = ev.Capacity > 0 && ev.ParticipantsCount >= ev.Capacity
		ev.Room.Visibility = "open"
		if isPrivate {
			ev.Room.Visibility = "closed"
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}
