package component

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

var (
	gray  = lipgloss.Color("240") // Gray color for general text
	red   = lipgloss.Color("196") // Red color for error messages
	pink  = lipgloss.Color("205") // Red color for error messages
	green = lipgloss.Color("34")  // Green color for success messages
	blue  = lipgloss.Color("33")  // Blue color for info messages
	white = lipgloss.Color("255") // White color for text

	// Styles for various UI components
	tabStyle = lipgloss.NewStyle().
			Padding(1, 1, 1, 1).
			Foreground(gray)
	activeTabStyle = tabStyle.
			Foreground(pink).
			Bold(true).
			Underline(true)
	keyListStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(gray)
	headerStyle = lipgloss.NewStyle().
			Padding(0, 1)
	headerLabelStyle = lipgloss.NewStyle().
				PaddingTop(1)
	TitleBarStyle = lipgloss.NewStyle().
			PaddingLeft(1)
	filterlabelStyle = lipgloss.NewStyle().PaddingLeft(1).Background(gray)
	filterFormStyle  = lipgloss.NewStyle().PaddingLeft(1)
)

func Form(label, value string, active bool, width int) string {
	form := lipgloss.NewStyle().
		Width(width / 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(gray)

	if active {
		form = form.BorderForeground(white)
	}

	return form.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		filterlabelStyle.Render(label+":"),
		filterFormStyle.Render(value),
	))

}

func labelAndName(label, name string) string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		headerLabelStyle.Render(label+":"),
		headerStyle.Render(name),
	)
}

func Header(host string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		" ",
		labelAndName("HOST", host),
	)
}

func ValueDisplay(value string, width int) string {
	return lipgloss.NewStyle().
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width).
		Render(value)
}

func TabRow(tabs int, currentTab int) string {
	_tabs := make([]string, tabs)
	for i := range tabs {
		if i == currentTab {
			_tabs[i] = activeTabStyle.Render(strconv.Itoa(i))
		} else {
			_tabs[i] = tabStyle.Render(strconv.Itoa(i))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, _tabs...)
}

func TitleBar(title string) lipgloss.Style {
	return TitleBarStyle.SetString(title)
}

func KeyList(keys []string, cur, height, width int) string {
	style := keyListStyle.Width(width)
	maxWidthKey := max(0, width-4)

	var keyFound = true
	if len(keys) == 0 {
		keys = []string{"No keys found."}
		keyFound = false
	}

	listItems := make([]string, max(len(keys), height))
	for i := range listItems {
		if i < len(keys) {
			if maxWidthKey > 3 && len(keys[i]) > maxWidthKey {
				listItems[i] = keys[i][:maxWidthKey-3] + "..." // Truncate long keys
			} else {
				listItems[i] = keys[i]
			}
		} else {
			listItems[i] = "" // Fill remaining space with empty strings
		}
	}

	l := list.New(listItems).
		ItemStyle(style).
		Enumerator(func(items list.Items, i int) string {
			if i == cur && keyFound {
				return "▶ " // Current item indicator
			}
			return ""
		}).
		ItemStyleFunc(func(items list.Items, i int) lipgloss.Style { return lipgloss.NewStyle() })

	return style.Render(l.String())
}

func HelpWindow() string {
	helpText := "Help:\n" +
		"  - j/k (or arrow keys): navigate key list\n" +
		"  - Enter: Update highlighted key\n" +
		"  - d: Delete highlighted key\n" +
		"  - /: Filter keys\n" +
		"  - n: Next key list\n" +
		"  - p: Previous key list\n" +
		"  - q or Esc or Ctrl-c:  quit\n"

	return lipgloss.NewStyle().
		Padding(1).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(gray).
		Render(helpText)
}
