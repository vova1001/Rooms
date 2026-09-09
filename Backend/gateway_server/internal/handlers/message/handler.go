package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	m "backend/gateway_server/models"
)

type Service interface {
	CreateMessage(ctx context.Context, roomID, userID, avatar, text string) (*m.Message, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(ctx context.Context, roomID, userID, avatar string, data json.RawMessage) (*m.Message, error) {
	var payload m.CreateMessagePayload

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal create message payload: %w", err)
	}

	text := strings.TrimSpace(payload.Text)

	if text == "" {
		return nil, fmt.Errorf("message text is empty")
	}

	message, err := h.service.CreateMessage(ctx, roomID, userID, avatar, text)

	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	return message, nil
}
