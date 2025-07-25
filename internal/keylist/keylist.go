package keylist

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/component"
	"github.com/hirotake111/redisclient/internal/values"
)

func Render(
	width int,
	height int,
	tabs int,
	currentTab int,
	keys []string,
	currentKeyIdx int,
	value values.Value,
	host string,
	filterHihghlighted bool,
	filterValue string,
	valueFormActive bool,
	updateFormValue string,
) string {
	// Calculate widths
	widthKeyListView := width / 3
	heightKeyListView := height - 10               // Adjust for header and footer
	heightValueView := height - 10                 // Adjust for header and footer
	widthValueView := width - widthKeyListView - 5 // Adjust for padding and borders

	tabRow := component.TabRow(tabs, currentTab)
	keyListTitle := component.TitleBarStyle.
		Width(widthKeyListView).
		Render("Keys")
	keyList := component.KeyList(keys, currentKeyIdx, heightKeyListView, widthKeyListView, !filterHihghlighted && !valueFormActive)
	keyListGroup := lipgloss.JoinVertical(lipgloss.Top, keyListTitle, keyList)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left,
			component.TitleBarStyle.Render("Value"),
			component.TTLIndicator(value.TTL()),
		),
		component.ValueDisplay(value.Data(), widthValueView, heightValueView),
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
