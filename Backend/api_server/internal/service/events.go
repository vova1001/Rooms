package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	m "backend/api_server/model"
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

	req.StartsAt = req.StartsAt.UTC()
	req.EndsAt = req.EndsAt.UTC()

	return s.repo.CreateEvent(ctx, req)
}
