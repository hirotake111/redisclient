package infoid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	idLength = 8
)

func New() (string, error) {
	b := make([]byte, idLength)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("Failed to generate a new ID: %w", err)
	}

	return hex.EncodeToString(b), nil
}
