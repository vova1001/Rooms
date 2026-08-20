package session

import (
	"crypto/sha256"
	"encoding/hex"
)

type HasherS struct{}

func NewHasher() *HasherS {
	return &HasherS{}
}

func (h *HasherS) Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
