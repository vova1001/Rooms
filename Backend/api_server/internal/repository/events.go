package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	m "backend/api_server/model"
)

var ErrRoomNotFound = errors.New("room not found")

func (r *PartRepo) CreateEvent(ctx context.Context, req m.CreateEventReq) (m.Event, error) {
	var ev m.Event

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO events (room_id, creator_id, title, description, starts_at, ends_at, price_minor)
		SELECT r.id, $2, $3, $4, $5, $6, $7
		FROM rooms r
		WHERE r.id = $1 AND r.owner_id = $2
		RETURNING id, room_id, creator_id, title, description, starts_at, ends_at, price_minor, currency, status, created_at`,
		req.RoomID, req.CreatorID, req.Title, req.Description, req.StartsAt, req.EndsAt, req.PriceMinor,
	).Scan(&ev.ID, &ev.RoomID, &ev.CreatorID, &ev.Title, &ev.Description,
		&ev.StartsAt, &ev.EndsAt, &ev.PriceMinor, &ev.Currency, &ev.Status, &ev.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.Event{}, ErrRoomNotFound
		}
		return m.Event{}, fmt.Errorf("create event: %w", err)
	}

	ev.IsFree = ev.PriceMinor == 0
	return ev, nil
}
