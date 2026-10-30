package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	repo "backend/api_server/internal/repository"
	"backend/api_server/internal/service/email"
	e "backend/api_server/internal/service/email"
	"backend/api_server/internal/service/otp"
	"backend/api_server/internal/service/session"
	m "backend/api_server/model"
	sharedRepoUsers "backend/shared/users"

	"github.com/google/uuid"
)

type PartService struct {
	sharedRepoUsers *sharedRepoUsers.RepositoryUser
	repo            *repo.PartRepo
	otp             *otp.Service
	otpRepo         otp.OTPRepository
	sender          email.Sender
	sessions        *session.Service
	limmiter        RateLimiter
}

var ErrValidation = errors.New("validation error")

func NewService(repo *repo.PartRepo, otp *otp.Service, otpRepo otp.OTPRepository, sender email.Sender, sessions *session.Service, limiter RateLimiter, sharedRepoUsers *sharedRepoUsers.RepositoryUser) *PartService {
	return &PartService{repo: repo, otpRepo: otpRepo, sender: sender, otp: otp, sessions: sessions, limmiter: limiter, sharedRepoUsers: sharedRepoUsers}
}

func (s *PartService) CreateUser(ctx context.Context, username, email, avatar string) (*m.User, error) {
	if username == "" {
		username = "user_" + uuid.New().String()[:8]
	}
	normEmail := e.Normalize(email)
	user, err := s.repo.CreateUser(ctx, username, normEmail, avatar)
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

	allowedIp, err := s.limmiter.AllowByIp(ctx, sec.IP)
	if err != nil {
		return fmt.Errorf("err limit Ip:%w", err)
	}

	if !allowedIp {
		return fmt.Errorf("TooManyRequestIp")
	}

	allowedEmail, err := s.limmiter.AllowByEmail(ctx, email)
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

func (s *PartService) VerifyCode(ctx context.Context, email, code string) (*m.VerifyCodeResult, error) {
	const (
		authSessionTTL         = 30 * 24 * time.Hour
		registrationSessionTTL = 30 * time.Minute
	)

	normEmail := e.Normalize(email)

	hash, err := s.otpRepo.GetOTP(ctx, normEmail)

	if err != nil {
		return nil, fmt.Errorf("err get old hash:%w", err)
	}

	allowed, err := s.limmiter.AllowVerifyByEmail(ctx, normEmail)
	if err != nil {
		return nil, fmt.Errorf("verify rate limit: %w", err)
	}
	if !allowed {
		return nil, fmt.Errorf("TooManyVerifyAttempts")
	}

	if !s.otp.Verify(normEmail, code, hash) {
		return nil, fmt.Errorf("err verify code:%w", err)
	}

	if err = s.otpRepo.DeleteOTP(ctx, normEmail); err != nil {
		return nil, fmt.Errorf("err delete otp in redis:%w", err)
	}

	user, err := s.repo.FindUserByEmail(ctx, normEmail)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if user != nil {
		token, err := s.sessions.CreateAuth(ctx, user.ID, authSessionTTL)

		if err != nil {
			return nil, fmt.Errorf("create auth session in s: %w", err)
		}

		return &m.VerifyCodeResult{Token: token, RequiresRegister: false}, nil
	}

	token, err := s.sessions.CreateRegistration(ctx, normEmail, registrationSessionTTL)
	if err != nil {
		return nil, fmt.Errorf("create registration session in s: %w", err)
	}

	return &m.VerifyCodeResult{Token: token, RequiresRegister: true}, nil
}

func (s *PartService) GetRegSession(ctx context.Context, token string) (string, error) {
	registerSession, err := s.sessions.GetRegistration(ctx, token)
	if err != nil {
		return "", fmt.Errorf("get reg session in s: %w", err)
	}

	return registerSession.Email, nil
}

func (s *PartService) CreateAuthSession(ctx context.Context, userId uuid.UUID) (string, error) {
	token, err := s.sessions.CreateAuth(ctx, userId, 30*24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("err create auth session in s: %w", err)
	}

	return token, nil
}

func (s *PartService) DeleteRegSession(ctx context.Context, token string) error {
	return s.sessions.DeleteRegistration(ctx, token)
}

func (s *PartService) GetAvatars(ctx context.Context) ([]repo.Avatars, error) {
	avatars, err := s.repo.GetAvatars(ctx)
	if err != nil {
		return nil, err
	}

	return avatars, nil
}

func (s *PartService) GetProfile(ctx context.Context, userId uuid.UUID) (m.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, userId)
	if err != nil {
		return m.Profile{}, fmt.Errorf("get profile: %w", err)
	}
	return profile, nil
}

func hasForbiddenRunes(s string, allowNewline bool) bool {
	for _, r := range s {
		if allowNewline && r == '\n' {
			continue
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return true
		}
	}
	return false
}

func (s *PartService) UpdateProfile(ctx context.Context, userId uuid.UUID, req m.UpdateProfileReq) (m.Profile, error) {
	if req.Username != nil {
		name := strings.TrimSpace(*req.Username)
		if name == "" || utf8.RuneCountInString(name) > 50 {
			return m.Profile{}, fmt.Errorf("%w: username must be 1-50 characters", ErrValidation)
		}
		if hasForbiddenRunes(name, false) {
			return m.Profile{}, fmt.Errorf("%w: username contains invalid characters", ErrValidation)
		}
		req.Username = &name
	}

	if req.Bio != nil {
		bio := strings.TrimSpace(*req.Bio)
		if utf8.RuneCountInString(bio) > 500 {
			return m.Profile{}, fmt.Errorf("%w: bio must be at most 500 characters", ErrValidation)
		}
		if hasForbiddenRunes(bio, true) {
			return m.Profile{}, fmt.Errorf("%w: bio contains invalid characters", ErrValidation)
		}
		req.Bio = &bio
	}

	if err := s.repo.ChangeProfile(ctx, userId, req); err != nil {
		return m.Profile{}, fmt.Errorf("Change profile: %w", err)
	}

	return s.repo.GetProfile(ctx, userId)
}
