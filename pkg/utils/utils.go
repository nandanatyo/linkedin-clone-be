package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

func GenerateRandomCode(length int) string {
	const charset = "0123456789"
	code := make([]byte, length)

	for i := range code {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[num.Int64()]
	}

	return string(code)
}

func GenerateOrderID() string {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	randomPart := GenerateRandomCode(6)
	return fmt.Sprintf("ORDER-%s-%s", timestamp, randomPart)
}
