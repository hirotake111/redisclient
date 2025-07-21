package keylist

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
)

func Render(
	width int,
	height int,
	tabs int,
	currentTab int,
	keys []string,
	currentKeyIdx int,
	value string,
	host string,
	filterHihghlighted bool,
	filterValue string,
) string {
	// Calculate widths
	widthKeyListView := width / 3
	heightKeyListView := height - 10                // Adjust for header and footer
	widthValueView := width - widthKeyListView - 10 // Adjust for padding and borders

	tabRow := component.TabRow(tabs, currentTab)
	keyListTitle := component.TitleBarStyle.
		Width(widthKeyListView).
		Render("Keys")
	keyList := component.KeyList(keys, currentKeyIdx, heightKeyListView, widthKeyListView)
	keyListGroup := lipgloss.JoinVertical(lipgloss.Top, keyListTitle, keyList)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		component.TitleBarStyle.Width(widthValueView).Render("Value"),
		component.ValueDisplay(value, widthValueView),
	)
	header := component.Header(host)
	filter := component.FilterForm("Filter", filterValue, filterHihghlighted, width)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabRow,
		filter,
		lipgloss.JoinHorizontal(lipgloss.Top,
			keyListGroup,
			valueDisplayGroup,
		),
	)
}
