package model

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component/helpbox"
	"github.com/hirotake111/redisclient/internal/keylist"
)

const (
	appShellPadding   = 2
	helpBoxHeight     = 3
	helpBoxPaddingTop = 1
)

var (
	appShellStyle = lipgloss.NewStyle().Padding(appShellPadding)
	helpBoxStyle  = lipgloss.NewStyle().PaddingTop(helpBoxPaddingTop)
)

func (m Model) View() string {
	width := m.width - appShellPadding*2
	height := m.height - appShellPadding*2 - helpBoxHeight - helpBoxPaddingTop
	helpBox := helpbox.New(helpBoxHeight)

	app := keylist.Render(
		m.mode,
		width,
		height,
		m.mode.Keys,
		m.HostName(),
		m.State,
	)

	return appShellStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			app,
			helpBoxStyle.Render(helpBox),
		),
	)
}
