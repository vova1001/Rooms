package email

import "context"

type Sender interface {
	SendCode(ctx context.Context, email, code string) error
}
