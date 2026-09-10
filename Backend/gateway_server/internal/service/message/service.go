package message

import (
	m "backend/gateway_server/models"
	"context"
	"fmt"
	"time"
)

type Repository interface {
	SaveMessage(ctx context.Context, msg *m.Message) (*m.Message, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMessage(ctx context.Context, roomID, userID, avatar, text string) (*m.Message, error) {
	msg := &m.Message{
		RoomID:    roomID,
		UserID:    userID,
		Avatar:    avatar,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}

	msg, err := s.repo.SaveMessage(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("create message in repository: %w", err)
	}

	return msg, nil
}
