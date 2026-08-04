package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	authSessionPrefix         = "session:auth:"
	registrationSessionPrefix = "session:registration:"
)

type Service struct {
	repo      RepositoryS
	generator Generator
	hasher    Hasher
}

func NewService(repo RepositoryS, generator Generator, hasher Hasher) *Service {
	return &Service{
		repo:      repo,
		generator: generator,
		hasher:    hasher,
	}
}

func (s *Service) CreateAuth(ctx context.Context, userId uuid.UUID, ttl time.Duration) (string, error) {
	sessionData := AuthSession{
		UserID: userId,
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("marshal registration session: %w", err)
	}

	token, err := s.generator.Generate()
	if err != nil {
		return "", fmt.Errorf("generate auth token: %w", err)
	}

	key := s.authKey(token)

	if err := s.repo.SaveSession(ctx, key, val, ttl); err != nil {
		return "", fmt.Errorf("save auth session: %w", err)
	}

	return token, nil
}

func (s *Service) CreateRegistration(ctx context.Context, email string, ttl time.Duration) (string, error) {
	sessionData := RegistrationSession{
		Email: email,
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		return "", fmt.Errorf("marshal registration session: %w", err)
	}

	token, err := s.generator.Generate()
	if err != nil {
		return "", fmt.Errorf("generate registration token: %w", err)
	}

	key := s.registrationKey(token)

	if err := s.repo.SaveSession(ctx, key, val, ttl); err != nil {
		return "", fmt.Errorf("save registration session: %w", err)
	}

	return token, nil
}

func (s *Service) GetAuth(ctx context.Context, token string) (*AuthSession, error) {

	key := s.authKey(token)
	val, err := s.repo.GetSession(ctx, key)
	if err != nil {
		return &AuthSession{}, fmt.Errorf("get auth session: %w", err)
	}

	var sessionData *AuthSession

	if err := json.Unmarshal(val, &sessionData); err != nil {
		return &AuthSession{}, fmt.Errorf("unmarshal auth session: %w", err)
	}

	return sessionData, nil
}

func (s *Service) GetRegistration(ctx context.Context, token string) (*RegistrationSession, error) {

	key := s.registrationKey(token)
	val, err := s.repo.GetSession(ctx, key)
	if err != nil {
		return &RegistrationSession{}, fmt.Errorf("get reg session: %w", err)
	}

	var sessionData RegistrationSession

	if err := json.Unmarshal(val, &sessionData); err != nil {
		return &RegistrationSession{}, fmt.Errorf("unmarshal reg session: %w", err)
	}

	return &sessionData, nil
}

func (s *Service) DeleteAuth(ctx context.Context, token string) error {

	key := s.authKey(token)
	if err := s.repo.DeleteSession(ctx, key); err != nil {
		return fmt.Errorf("unmarshal auth session: %w", err)
	}

	return nil
}

func (s *Service) DeleteRegistration(ctx context.Context, token string) error {

	key := s.registrationKey(token)
	if err := s.repo.DeleteSession(ctx, key); err != nil {
		return fmt.Errorf("unmarshal reg session: %w", err)
	}

	return nil
}

func (s *Service) authKey(token string) string {
	tokenHash := s.hasher.Hash(token)

	return authSessionPrefix + tokenHash
}

func (s *Service) registrationKey(token string) string {
	tokenHash := s.hasher.Hash(token)

	return registrationSessionPrefix + tokenHash
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
