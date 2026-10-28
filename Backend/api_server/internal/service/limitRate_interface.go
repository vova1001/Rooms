package service

import "context"

type RateLimiter interface {
	AllowByIp(ctx context.Context, ip string) (bool, error)
	AllowByEmail(ctx context.Context, email string) (bool, error)
	AllowVerifyByEmail(ctx context.Context, email string) (bool, error)
}
