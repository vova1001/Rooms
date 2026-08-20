package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type CodeGenerator struct{}

func NewGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

func (g *CodeGenerator) Generate() (string, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("err generate:%w", err)
	}

	return fmt.Sprintf("%06d", num.Int64()), nil
}
