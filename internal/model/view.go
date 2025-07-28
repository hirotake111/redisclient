package model

import (
	"fmt"

	"github.com/hirotake111/redisclient/internal/help"
	"github.com/hirotake111/redisclient/internal/keylist"
)

func (m Model) View() string {
	switch m.state {
	case ListState:
		return keylist.Render(
			m.width,
			m.height,
			m.tabs,
			m.currentTab,
			m.CurrentKeyList(),
			m.currentKeyIdx,
			m.value,
			m.HostName(),
			m.errorMsg,
			m.keyHistoryIdx,
			int(m.redisCursor),
			m.filterForm,
			m.updateForm,
		)

	case HelpWindowState:
		return help.Render(m.width, m.height)
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
}
