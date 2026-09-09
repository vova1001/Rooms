package getauthsession

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	authSessionPrefix = "session:auth:"
)

type AuthSession struct {
	UserID uuid.UUID `json:"user_id"`
}

type Repository interface {
	GetSession(ctx context.Context, key string) ([]byte, error)
}

type Hasher interface {
	Hash(token string) string
}

type Reader struct {
	repo   Repository
	hasher Hasher
}

func NewReader(repo Repository, hasher Hasher) *Reader {
	return &Reader{repo: repo, hasher: hasher}
}

func (r *Reader) GetAuth(ctx context.Context, token string) (*AuthSession, error) {

	key := authSessionPrefix + r.hasher.Hash(token)
	val, err := r.repo.GetSession(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get auth session: %w", err)
	}

	var sessionData AuthSession

	if err := json.Unmarshal(val, &sessionData); err != nil {
		return nil, fmt.Errorf("unmarshal auth session: %w", err)
	}

	return &sessionData, nil
}
