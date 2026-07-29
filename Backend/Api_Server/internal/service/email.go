package service

import (
	"fmt"
	"net/mail"
	"strings"
)

func Normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func Validate(email string) error {
	if email == "" {
		return fmt.Errorf("email empty")
	}

	if len(email) > 254 {
		return fmt.Errorf("email too long")
	}

	address, err := mail.ParseAddress(email)

	if err != nil {
		return fmt.Errorf("invalid email")
	}

	if address.Address != email {
		return fmt.Errorf("invalid email")
	}
	return nil

}
