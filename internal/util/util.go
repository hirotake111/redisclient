package util

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/command"
)

func LogMsg(prefix string, msg tea.Msg) {
	if m, ok := msg.(command.MsgWithKind); ok {
		log.Printf("%s - KIND: %s", prefix, m.Kind())
	} else {
		log.Printf("%s - %+v", prefix, msg)
	}
}
