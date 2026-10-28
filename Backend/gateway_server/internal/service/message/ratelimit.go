package message

import (
	"context"
	"errors"
	"time"
)

const (
	messageRateLimit  = 10
	messageRateWindow = 5 * time.Second
)

var ErrRateLimited = errors.New("message rate limit exceeded")

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

func messageRateKey(userID string) string {
	return "rate_limit:chat:message:" + userID
}
