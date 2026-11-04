package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	m "backend/api_server/model"

	"github.com/google/uuid"
)

func (s *PartService) CreateEvent(ctx context.Context, req m.CreateEventReq) (m.Event, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if n := utf8.RuneCountInString(req.Title); n == 0 || n > 200 {
		return m.Event{}, fmt.Errorf("%w: title must be 1-200 characters", ErrValidation)
	}
	if utf8.RuneCountInString(req.Description) > 2000 {
		return m.Event{}, fmt.Errorf("%w: description must be at most 2000 characters", ErrValidation)
	}
	if req.StartsAt.IsZero() || req.EndsAt.IsZero() {
		return m.Event{}, fmt.Errorf("%w: starts_at and ends_at are required", ErrValidation)
	}
	if req.StartsAt.Before(time.Now()) {
		return m.Event{}, fmt.Errorf("%w: starts_at must be in the future", ErrValidation)
	}
	if !req.EndsAt.After(req.StartsAt) {
		return m.Event{}, fmt.Errorf("%w: ends_at must be after starts_at", ErrValidation)
	}
	if req.EndsAt.Sub(req.StartsAt) > 12*time.Hour {
		return m.Event{}, fmt.Errorf("%w: event can last at most 12 hours", ErrValidation)
	}

	if req.IsFree {
		req.PriceMinor = 0
	} else if req.PriceMinor <= 0 || req.PriceMinor > 100_000_000 {
		return m.Event{}, fmt.Errorf("%w: price_minor must be between 1 and 100000000 for a paid event", ErrValidation)
	}

	if req.Capacity < 0 || req.Capacity > 1000 {
		return m.Event{}, fmt.Errorf("%w: capacity must be between 0 and 1000 (0 = unlimited)", ErrValidation)
	}

	req.StartsAt = req.StartsAt.UTC()
	req.EndsAt = req.EndsAt.UTC()

	return s.repo.CreateEvent(ctx, req)
}

func (s *PartService) ListEvents(ctx context.Context, date, timeZone, roomID string) ([]m.EventItem, error) {
	if timeZone == "" {
		timeZone = "UTC"
	}
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid time_zone", ErrValidation)
	}

	dayStart, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, fmt.Errorf("%w: date must be YYYY-MM-DD", ErrValidation)
	}
	dayEnd := dayStart.AddDate(0, 0, 1)

	var room uuid.NullUUID
	if roomID != "" {
		id, err := uuid.Parse(roomID)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid room_id", ErrValidation)
		}
		room = uuid.NullUUID{UUID: id, Valid: true}
	}

	return s.repo.ListEvents(ctx, dayStart, dayEnd, room)
}

func (s *PartService) JoinEvent(ctx context.Context, eventID, userID uuid.UUID) (m.Booking, error) {
	return s.repo.JoinEvent(ctx, eventID, userID)
}

func (s *PartService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]m.MyBooking, error) {
	return s.repo.GetUserBookings(ctx, userID)
}

func (s *PartService) GetOwnerEvents(ctx context.Context, ownerID uuid.UUID) ([]m.OwnerEvent, error) {
	return s.repo.GetOwnerEvents(ctx, ownerID)
}

func (s *PartService) DecideBooking(ctx context.Context, bookingID, ownerID uuid.UUID, status string) (m.Booking, error) {
	if status != "confirmed" && status != "rejected" {
		return m.Booking{}, fmt.Errorf("%w: status must be confirmed or rejected", ErrValidation)
	}
	return s.repo.DecideBooking(ctx, bookingID, ownerID, status)
}
