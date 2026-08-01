package otp

import "context"

type OTPRepository interface {
	SaveOTP(ctx context.Context, email string, hash string) error

	GetOTP(ctx context.Context, email string) (string, error)

	DeleteOTP(ctx context.Context, email string) error
}

type Generator interface {
	Generate() (string, error)
}

type Hasher interface {
	Hash(email, code string) string
	Verify(email, code, hash string) bool
}
