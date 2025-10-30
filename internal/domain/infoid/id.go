package infoid

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	idLength = 8
)

func New() (string, error) {
	b := make([]byte, idLength)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
