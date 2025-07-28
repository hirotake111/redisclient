package keylist

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
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
	errorMsg string,
	page int,
	cursor int,
	filterForm *textarea.Model,
	updateForm *textarea.Model,
) string {
	// Calculate widths
	widthKeyListView := width / 3
	heightLeftPane := height - 10  // Adjust for header and footer
	heightRightPane := height - 10 // Adjust for header and footer
	heightValueDisplay := heightRightPane - 5
	heightErrorBox := 3                            // Space for error messages
	widthRightPane := width - widthKeyListView - 5 // Adjust for padding and borders

	tabRow := component.TabRow(tabs, currentTab)
	keyListTitle := component.TitleBarStyle.
		Width(widthKeyListView).
		Render(fmt.Sprintf("Keys (page: %d, cursor: %d)", page, cursor))
	keyList := component.KeyList(keys, currentKeyIdx, heightLeftPane, widthKeyListView, !filterHihghlighted && !valueFormActive)
	keyListGroup := lipgloss.JoinVertical(lipgloss.Top, keyListTitle, keyList)

	valueDisplayGroup := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Left,
			component.TitleBarStyle.Render("Value"),
			component.TTLIndicator(value.TTL()),
		),
		component.ValueDisplay(value.Data(), widthRightPane, heightValueDisplay),
		component.ErrorBox(errorMsg, widthRightPane, heightErrorBox),
	)
	header := component.Header(host)

	var form string
	if updateForm.Focused() {
		form = updateForm.View()
	} else {
		form = filterForm.View()
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
