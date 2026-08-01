package session

import (
	"context"
	"time"
)

type RepositoryS interface {
	Save(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type Generator interface {
	Generate() (string, error)
}

type Hasher interface {
	Hash(token string) string
}
