package internal

import (
	m "Signal_Server/Models"

	"context"
	"fmt"
)

type UserRepo interface {
	AddUser(ctx context.Context, RoomId string, user *m.User) error
}

type Service struct {
	repo *repoPart
}

func NewService(repo *repoPart) *Service {
	return &Service{repo: repo}
}

func (s Service) Join(ctx context.Context, roomId string, user *m.User) (*m.User, error) {
	if err := s.repo.AddUser(ctx, roomId, user); err != nil {
		return nil, fmt.Errorf("Error into repoPart addUser:%w", err)
	}
	return user, nil
}
