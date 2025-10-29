package keylist

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
	"github.com/hirotake111/redisclient/internal/mode"
	"github.com/hirotake111/redisclient/internal/state"
)

const (
	heightErrorBox = 5 // Space for error messages
)

func Render(
	mode *mode.ListMode,
	width int,
	height int,
	keys []string,
	host string,
	st state.AppState,
) string {
	// Calculate widths and heights
	heightValueDisplay := height - heightErrorBox - 7
	heightLeftPane := heightValueDisplay + heightErrorBox + 2
	widthLeftPane := width / 3
	widthRightPane := width - widthLeftPane - 5

	tabRow := component.TabRow(mode.Tabs, mode.CurrentTab)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		mode.Viewport.View(widthRightPane, heightValueDisplay, st),
		component.ErrorBox(mode.ErrorMsg, widthRightPane, heightErrorBox),
	)

	hostHeader := component.HostHeader(host)

	return lipgloss.JoinVertical(lipgloss.Left,
		tabRow,
		lipgloss.JoinHorizontal(lipgloss.Top,
			mode.KeyList.View(widthLeftPane, heightLeftPane, st),
			valueDisplayGroup,
		),
		hostHeader,
	)
}
