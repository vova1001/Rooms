package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const tokenSize = 32

type GeneratorS struct{}

func NewGeneratorS() *GeneratorS {
	return &GeneratorS{}
}

func (g *GeneratorS) Generate() (string, error) {
	randomBytes := make([]byte, tokenSize)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(randomBytes)

	return token, nil
}
