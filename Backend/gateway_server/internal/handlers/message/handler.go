package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	m "backend/gateway_server/models"
)

type Service interface {
	CreateMessage(ctx context.Context, roomID, userID, userName, avatar, text string) (*m.Message, error)
	GetHistory(ctx context.Context, roomID string, before *time.Time) ([]m.Message, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(ctx context.Context, roomID, userID, userName, avatar string, data json.RawMessage) (*m.Message, error) {
	var payload m.CreateMessagePayload

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal create message payload: %w", err)
	}

	text := strings.TrimSpace(payload.Text)

	if text == "" {
		return nil, fmt.Errorf("message text is empty")
	}

	message, err := h.service.CreateMessage(ctx, roomID, userID, userName, avatar, text)

	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}

func (h *Handler) History(ctx context.Context, roomID string, data json.RawMessage) ([]m.Message, error) {
	var payload m.HistoryPayload

	if len(data) > 0 {
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unmarshal history payload: %w", err)
		}
	}

	messages, err := h.service.GetHistory(ctx, roomID, payload.Before)
	if err != nil {
		return nil, fmt.Errorf("history: %w", err)
	}

	return messages, nil
}
