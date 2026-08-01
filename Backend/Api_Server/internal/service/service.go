package service

import (
	"context"
	"errors"
	"fmt"

	repo "rooms/internal/repository"
	"rooms/internal/service/email"
	e "rooms/internal/service/email"
	"rooms/internal/service/otp"
	m "rooms/model"

	"github.com/google/uuid"
)

type PartService struct {
	repo    *repo.PartRepo
	otp     *otp.Service
	otpRepo otp.OTPRepository
	sender  email.Sender
}

func NewService(repo *repo.PartRepo, otp *otp.Service, otpRepo otp.OTPRepository, sender email.Sender) *PartService {
	return &PartService{repo: repo, otpRepo: otpRepo, sender: sender, otp: otp}
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

func (s *PartService) SendCodeFromEmail(ctx context.Context, sec m.SendEmailCodeRequest) error {
	email := e.Normalize(sec.Email)

	if err := e.Validate(email); err != nil {
		return fmt.Errorf("err validate email: %w", err)
	}

	allowedIp, err := s.repo.AllowByIp(ctx, sec.IP)
	if err != nil {
		return fmt.Errorf("err limit Ip:%w", err)
	}

	if !allowedIp {
		return fmt.Errorf("TooManyRequestIp")
	}

	allowedEmail, err := s.repo.AllowByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("err limit Emailt:%w", err)
	}

	if !allowedEmail {
		return fmt.Errorf("TooManyRequestEmail")
	}

	code, err := s.otp.Generate()
	if err != nil {
		return fmt.Errorf("err generate code:%w", err)
	}

	hash := s.otp.Hash(email, code)

	if err = s.otpRepo.SaveOTP(ctx, email, hash); err != nil {
		return fmt.Errorf("err save hash to redis:%w", err)
	}

	if err = s.sender.SendCode(ctx, email, code); err != nil {
		return fmt.Errorf("err send code:%w", err)
	}

	return nil
}

func (s *PartService) VerifyCode(ctx context.Context, email, code string) error {
	hash, err := s.otpRepo.GetOTP(ctx, email)
	if err != nil {
		return fmt.Errorf("err get old hash:%w", err)
	}

	if !s.otp.Verify(email, code, hash) {
		return fmt.Errorf("err verify code:%w", err)
	}

	if err = s.otpRepo.DeleteOTP(ctx, email); err != nil {
		return fmt.Errorf("err delete otp in redis:%w", err)
	}
}
