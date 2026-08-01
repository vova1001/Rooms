package repository

import (
	"context"
	"time"
)

const otpTTL = 5 * time.Minute

func (r *PartRepo) SaveOTP(ctx context.Context, email, hash string) error {
	return r.rdb.Set(ctx, "otp"+email, hash, otpTTL).Err()
}

func (r *PartRepo) GetOTP(ctx context.Context, email string) (string, error) {
	return r.rdb.Get(ctx, "otp"+email).Result()
}

func (r *PartRepo) DeleteOTP(ctx context.Context, email string) error {
	return r.rdb.Del(ctx, "otp"+email).Err()
}
