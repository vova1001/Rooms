package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRoomCode = errors.New("invalid room code")
	ErrTooManyAttempts = errors.New("too many attempts")
)

func (s *PartService) UnlockRoom(ctx context.Context, roomID, userID uuid.UUID, password string) error {
	isPrivate, ownerID, passHash, err := s.repo.GetRoomLock(ctx, roomID)
	if err != nil {
		return err
	}

	if isPrivate && ownerID != userID {
		allowed, err := s.repo.AllowRoomUnlock(ctx, roomID, userID)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrTooManyAttempts
		}

		if !passHash.Valid || bcrypt.CompareHashAndPassword([]byte(passHash.String), []byte(password)) != nil {
			return ErrInvalidRoomCode
		}
	}

	if err := s.repo.AddUserToRoom(ctx, roomID, userID); err != nil {
		return fmt.Errorf("add user to room: %w", err)
	}
	return nil
}
