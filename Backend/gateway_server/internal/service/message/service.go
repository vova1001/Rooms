package message

import (
	m "backend/gateway_server/models"
	"context"
	"fmt"
	"time"
)

type Repository interface {
	SaveMessage(ctx context.Context, msg *m.Message) (*m.Message, error)
	GetMessages(ctx context.Context, roomID string, before *time.Time, limit int) ([]m.Message, error)
}

type Service struct {
	repo    Repository
	limiter RateLimiter
}

func NewService(repo Repository, limiter RateLimiter) *Service {
	return &Service{repo: repo, limiter: limiter}
}

func (s *Service) CreateMessage(ctx context.Context, roomID, userID, userName, avatar, text string) (*m.Message, error) {
	allowed, err := s.limiter.Allow(ctx, messageRateKey(userID), messageRateLimit, messageRateWindow)
	if err != nil {
		return nil, fmt.Errorf("check message rate limit: %w", err)
	}
	if !allowed {
		return nil, ErrRateLimited
	}

	msg := &m.Message{
		RoomID:    roomID,
		UserID:    userID,
		UserName:  userName,
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

const (
	historyLimit = 50
)

func (s *Service) GetHistory(ctx context.Context, roomID string, before *time.Time) ([]m.Message, error) {
	messages, err := s.repo.GetMessages(ctx, roomID, before, historyLimit)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	return messages, nil
}
