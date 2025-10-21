package keylist

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
	"github.com/hirotake111/redisclient/internal/mode"
)

func Render(
	mode *mode.ListMode,
	width int,
	height int,
	keys []string,
	host string,
) string {
	// Reserve 2 lines for the help pane (1 for help, 1 for spacing)
	helpPaneHeight := 2
	height = max(0, height-helpPaneHeight)
	// Calculate widths and heights
	widthKeyListView := width / 3
	heightLeftPane := height - 10  // Adjust for header and footer
	heightRightPane := height - 10 // Adjust for header and footer
	heightValueDisplay := heightRightPane - 5
	heightErrorBox := 3                            // Space for error messages
	widthRightPane := width - widthKeyListView - 5 // Adjust for padding and borders

	tabRow := component.TabRow(mode.Tabs, mode.CurrentTab)
	keyListTitle := component.TitleBarStyle.
		Width(widthKeyListView).
		Render(fmt.Sprintf("Keys (page: %d, cursor: %d)", mode.KeyHistoryIdx, mode.RedisCursor))
	keyList := component.KeyList(keys, mode.CurrentKeyIdx, heightLeftPane, widthKeyListView, !(mode.FilterForm.Focused() || mode.UpdateForm.Focused()))
	keyListGroup := lipgloss.JoinVertical(lipgloss.Top, keyListTitle, keyList)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left,
			component.TitleBarStyle.Render("Value"),
			component.TTLIndicator(mode.Value.TTL()),
		),
		component.ValueDisplay(mode.Value.Data(), widthRightPane, heightValueDisplay),
		component.ErrorBox(mode.ErrorMsg, widthRightPane, heightErrorBox),
	)
	header := component.HostHeader(host)

	var form string
	if mode.UpdateForm.Focused() {
		form = mode.UpdateForm.View()
	} else {
		form = mode.FilterForm.View()
	}

	app := lipgloss.JoinVertical(lipgloss.Left,
		tabRow,
		form,
		lipgloss.JoinHorizontal(lipgloss.Top,
			keyListGroup,
			valueDisplayGroup,
		),
		header,
	)

	// Add the always-visible help pane at the bottom
	return lipgloss.JoinVertical(lipgloss.Left, app, component.HelpPane())
}
