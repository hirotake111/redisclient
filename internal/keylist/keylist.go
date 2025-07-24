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
	valueFormActive bool,
	updateFormValue string,
) string {
	// Calculate widths
	widthKeyListView := width / 3
	heightKeyListView := height - 10                // Adjust for header and footer
	heightValueView := height - 10                  // Adjust for header and footer
	widthValueView := width - widthKeyListView - 10 // Adjust for padding and borders

	tabRow := component.TabRow(tabs, currentTab)
	keyListTitle := component.TitleBarStyle.
		Width(widthKeyListView).
		Render("Keys")
	keyList := component.KeyList(keys, currentKeyIdx, heightKeyListView, widthKeyListView)
	keyListGroup := lipgloss.JoinVertical(lipgloss.Top, keyListTitle, keyList)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		component.TitleBarStyle.Width(widthValueView).Render("Value"),
		component.ValueDisplay(value, widthValueView, heightValueView),
	)
	header := component.Header(host)

	var form string
	if valueFormActive {
		form = component.Form("New Value", updateFormValue, true, width)
	} else {
		form = component.Form("Filter", filterValue, filterHihghlighted, width)
	}

	app := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabRow,
		form,
		lipgloss.JoinHorizontal(lipgloss.Top,
			keyListGroup,
			valueDisplayGroup,
		),
	)

	return app
}
