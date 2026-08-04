package session

import "github.com/google/uuid"

type AuthSession struct {
	UserID uuid.UUID `json:"user_id"`
}

type RegistrationSession struct {
	Email string `json:"email"`
}
