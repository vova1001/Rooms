package service

import (
	"context"
	"errors"
	"fmt"

	repo "rooms/internal/repository"
	m "rooms/model"

	"github.com/google/uuid"
)

type PartService struct {
	repo *repo.PartRepo
}

func NewService(repo *repo.PartRepo) *PartService {
	return &PartService{repo: repo}
}

func (s *PartService) CreateUser(ctx context.Context, username, email, avatar string) (*m.User, error) {
	if username == "" {
		username = "user_" + uuid.New().String()[:8]
	}
	user, err := s.repo.CreateUser(ctx, username, email, avatar)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}
	return user, nil
}

func (s *PartService) CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*m.Room, error) {
	if name == "" {
		return nil, errors.New("room name cannot be empty")
	}
	if ownerID == uuid.Nil {
		return nil, errors.New("owner id is required")
	}
	room, err := s.repo.CreateRoom(ctx, name, ownerID)
	if err != nil {
		return nil, fmt.Errorf("CreateRoom: %w", err)
	}
	// Добавляем владельца в комнату (связка room_users)
	if err := s.repo.AddUserToRoom(ctx, room.ID, ownerID); err != nil {
		// Здесь можно либо вернуть ошибку, либо залогировать. По заданию – вернём ошибку.
		return nil, fmt.Errorf("CreateRoom: failed to add owner to room: %w", err)
	}
	return room, nil
}

func (s *PartService) GetAllRooms(ctx context.Context) ([]*m.Room, error) {
	rooms, err := s.repo.GetAllRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllRooms: %w", err)
	}
	return rooms, nil
}

func (s *PartService) GetRoomUsers(ctx context.Context, roomID uuid.UUID) ([]*m.User, error) {
	if roomID == uuid.Nil {
		return nil, errors.New("GetRoomUsers: room id is required")
	}
	users, err := s.repo.GetUsersByRoomID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("GetRoomUsers: %w", err)
	}
	return users, nil
}

func (s *PartService) SendCode(ctx context.Context, sec m.SendEmailCodeRequest) error {
	email := Normalize(sec.Email)

	if err := Validate(email); err != nil {
		return fmt.Errorf("err validate email: %w", err)
	}

	allowedIp, err := s.repo.AllowByIp(ctx, sec.IP)
	if err != nil {
		return fmt.Errorf("err limit Ip:%w", err)
	}

	if !allowedIp {
		return fmt.Errorf("TooManyRequestIp")
	}

	allowedEmail, err := s.repo.AllowByEmail(ctx, sec.IP)
	if err != nil {
		return fmt.Errorf("err limit Emailt:%w", err)
	}

	if !allowedEmail {
		return fmt.Errorf("TooManyRequestEmail")
	}

}
