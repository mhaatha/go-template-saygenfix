package helper

import (
	"crypto/rand"
	"fmt"
	"io"
)

const DefaultExamIDLength = 8

// GenerateSecureString membuat ID acak yang aman.
func GenerateSecureString() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, DefaultExamIDLength)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("could not generate random bytes: %w", err)
	}

	for i := 0; i < DefaultExamIDLength; i++ {
		// Ambil byte acak dan petakan ke dalam charset
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}
