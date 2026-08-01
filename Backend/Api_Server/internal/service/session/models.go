package session

type AuthSession struct {
	UserID int64 `json:"user_id"`
}

type RegistrationSession struct {
	Email string `json:"email"`
}
