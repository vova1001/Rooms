package repository

import (
	"context"
	"fmt"
	"time"
)

func (r *PartRepo) SaveSession(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if err := r.rdb.Set(ctx, key, val, ttl).Err(); err != nil {
		return fmt.Errorf("redis save session: %w", err)
	}

	return nil
}

func (r *PartRepo) GetSession(ctx context.Context, key string) ([]byte, error) {
	res, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("redis get session: %w", err)
	}

	return res, nil
}

func (r *PartRepo) DeleteSession(ctx context.Context, key string) error {
	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete session: %w", err)
	}
	return nil
}
