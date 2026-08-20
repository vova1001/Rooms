package otp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type CodeHasher struct {
	secret []byte
}

func NewCodeHasher(secret string) *CodeHasher {
	return &CodeHasher{secret: []byte(secret)}
}

func (h *CodeHasher) Hash(email, code string) string {
	mac := hmac.New(sha256.New, h.secret)

	mac.Write([]byte(email))
	mac.Write([]byte{0})
	mac.Write([]byte(code))

	return hex.EncodeToString(mac.Sum(nil))
}

func (h *CodeHasher) Verify(email, code, hash string) bool {
	newHash := h.Hash(email, code)

	return hmac.Equal([]byte(newHash), []byte(hash))
}
