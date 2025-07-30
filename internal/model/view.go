package model

import (
	"fmt"

	"github.com/hirotake111/redisclient/internal/help"
	"github.com/hirotake111/redisclient/internal/keylist"
)

func (m Model) View() string {
	switch m.State {
	case ListState:
		return keylist.Render(
			m.mode,
			m.width,
			m.height,
			m.CurrentKeyList(),
			m.HostName(),
		)

	case HelpWindowState:
		return help.Render(m.width, m.height)
	}

	return fmt.Sprintf("Unknown state: %s", m.State)
}
