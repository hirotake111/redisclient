package util

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/command"
)

const (
	idLength = 8
)

func NewID() (string, error) {
	b := make([]byte, idLength)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func LogMsg(prefix string, msg tea.Msg) {
	if m, ok := msg.(command.MsgWithKind); ok {
		log.Printf("%s - KIND: %s", prefix, m.Kind())
	} else {
		log.Printf("%s - %+v", prefix, msg)
	}
}
