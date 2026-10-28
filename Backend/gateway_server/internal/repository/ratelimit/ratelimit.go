package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var incrScript = redis.NewScript(`
	local count = redis.call("INCR", KEYS[1])
	if count == 1 then
		redis.call("PEXPIRE", KEYS[1], ARGV[1])
	end
	return count
`)

type Repository struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Repository {
	return &Repository{rdb: rdb}
}

func (r *Repository) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := incrScript.Run(ctx, r.rdb, []string{key}, window.Milliseconds()).Int()
	if err != nil {
		return false, fmt.Errorf("rate limit key=%s: %w", key, err)
	}

	return count <= limit, nil
}
