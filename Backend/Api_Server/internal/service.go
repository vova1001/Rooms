package internal

import (
	"context"
	"errors"
	"fmt"

	m "rooms/model"

	"github.com/google/uuid"
)

type partService struct {
	repo *PartRepo
}

func NewService(repo *PartRepo) *partService {
	return &partService{repo: repo}
}

func (s *partService) CreateUser(ctx context.Context, username string) (*m.User, error) {
	if username == "" {
		username = "user_" + uuid.New().String()[:8]
	}
	user, err := s.repo.CreateUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: %w", err)
	}
	return user, nil
}

func (s *partService) CreateRoom(ctx context.Context, name string, ownerID uuid.UUID) (*m.Room, error) {
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

func (s *partService) GetAllRooms(ctx context.Context) ([]*m.Room, error) {
	rooms, err := s.repo.GetAllRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllRooms: %w", err)
	}
	return rooms, nil
}

func (s *partService) GetRoomUsers(ctx context.Context, roomID uuid.UUID) ([]*m.User, error) {
	if roomID == uuid.Nil {
		return nil, errors.New("GetRoomUsers: room id is required")
	}
	users, err := s.repo.GetUsersByRoomID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("GetRoomUsers: %w", err)
	}
	return users, nil
}
