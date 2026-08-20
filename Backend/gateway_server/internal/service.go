package internal

import (
	"backend/gateway_server/livekit"
	m "backend/gateway_server/models"
	"time"

	"context"
	"fmt"
)

type UserRepo interface {
	AddUser(ctx context.Context, RoomId string, user *m.User) error
}

type Service struct {
	repo *repoPart
	LKS  *livekit.TokenService
}

type JoinResult struct {
	RoomID         string                  `json:"room_id"`
	ConnectionData *livekit.ConnectionData `json:"livekit"`
}

func NewService(repo *repoPart, LKS *livekit.TokenService) *Service {
	return &Service{
		repo: repo,
		LKS:  LKS,
	}
}

func (s Service) Join(ctx context.Context, roomId string, user *m.User) (*JoinResult, error) {
	user.CreatedAt = time.Now().UTC()
	if err := s.repo.AddUser(ctx, roomId, user); err != nil {
		return nil, fmt.Errorf("Error into repoPart addUser:%w", err)
	}

	res, err := s.LKS.CreateConnectionData(roomId, user.Id, user.UserName)
	if err != nil {
		return nil, fmt.Errorf("err create connection LKS: %w", err)
	}

	return &JoinResult{
		RoomID:         roomId,
		ConnectionData: res,
	}, nil
}

func (s Service) Leave(ctx context.Context, roomId, userId string) error {
	if err := s.repo.DeleteUser(ctx, roomId, userId); err != nil {
		return err
	}
	return nil
}
