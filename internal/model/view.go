package model

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
	"github.com/hirotake111/redisclient/internal/component/helpbox"
)

const (
	appShellPadding   = 2
	helpBoxHeight     = 3
	helpBoxPaddingTop = 1
	heightErrorBox    = 5 // Space for error messages
)

var (
	appShellStyle = lipgloss.NewStyle().Padding(appShellPadding)
	helpBoxStyle  = lipgloss.NewStyle().PaddingTop(helpBoxPaddingTop)
)

func (m Model) View() string {
	width := m.width - appShellPadding*2
	height := m.height - appShellPadding*2 - helpBoxHeight - helpBoxPaddingTop

	heightValueDisplay := height - heightErrorBox - 7
	heightLeftPane := heightValueDisplay + heightErrorBox + 2
	widthLeftPane := width / 3
	widthRightPane := width - widthLeftPane - 5

	// Help box
	helpBox := helpBoxStyle.Render(helpbox.New(helpBoxHeight))

	// Database tab
	tab := component.TabRow(m.tabs, m.currentTab)

	// Message box
	msgbox := component.ErrorBox(m.errorMsg, widthRightPane, heightErrorBox)

	// Key list
	left := m.keyList.View(widthLeftPane, heightLeftPane, m.State)

	// Viewport
	viewport := m.viewport.View(widthRightPane, heightValueDisplay, m.State)

	// Connection display
	bottom := component.HostHeader(m.HostName())

	// Right pane
	right := lipgloss.JoinVertical(lipgloss.Top, viewport, msgbox)

	middle := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	main := lipgloss.JoinVertical(lipgloss.Left, tab, middle, bottom, helpBox)

	return appShellStyle.Render(main)
}
