package keylist

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
	"github.com/hirotake111/redisclient/internal/mode"
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
) string {
	// Calculate widths and heights
	heightValueDisplay := height - heightErrorBox - 10
	heightLeftPane := heightValueDisplay + heightErrorBox + 3
	widthLeftPane := width / 3
	widthRightPane := width - widthLeftPane - 5

	tabRow := component.TabRow(mode.Tabs, mode.CurrentTab)

	keyListHighlighted := !(mode.FilterForm.Focused() || mode.UpdateForm.Focused())
	keyListGroup := lipgloss.JoinVertical(lipgloss.Top,
		component.KeyListTitle(widthLeftPane, mode.KeyHistoryIdx, mode.RedisCursor),
		component.KeyList(keys, mode.CurrentKeyIdx, heightLeftPane, widthLeftPane, keyListHighlighted),
	)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		component.ValueTitle(mode.Value.TTL()),
		component.ValueDisplay(mode.Value.Data(), widthRightPane, heightValueDisplay),
		component.ErrorTitle(widthLeftPane),
		component.ErrorBox(mode.ErrorMsg, widthRightPane, heightErrorBox),
	)

	hostHeader := component.HostHeader(host)

	var form string
	if mode.UpdateForm.Focused() {
		form = mode.UpdateForm.View()
	} else {
		form = mode.FilterForm.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		tabRow,
		form,
		lipgloss.JoinHorizontal(lipgloss.Top,
			keyListGroup,
			valueDisplayGroup,
		),
		hostHeader,
	)
}
