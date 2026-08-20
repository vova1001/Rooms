package session

import (
	"context"
	"time"
)

type RepositoryS interface {
	SaveSession(ctx context.Context, key string, value []byte, ttl time.Duration) error
	GetSession(ctx context.Context, key string) ([]byte, error)
	DeleteSession(ctx context.Context, key string) error
}

type Generator interface {
	Generate() (string, error)
}

type Hasher interface {
	Hash(token string) string
}
