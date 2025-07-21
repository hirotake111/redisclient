package component

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

const (
	tl  = "╭" // Top left corner for key list
	tr  = "╮" // Top right corner for key list
	bl  = "╰" // Bottom left corner for key list
	br  = "╯" // Bottom right corner for key list
	hl  = "─" // Horizontal line for key list
	vl  = "│" // Vertical line for key list
	dhl = "═" // Double horizontal line for key list
	dvl = "║" // Double vertical line for key list
	tld = "╔" // Top left double corner for key list
	trd = "╗" // Top right double corner for key list
	bld = "╚" // Bottom left double corner for key list
	brd = "╝" // Bottom right double corner for key list
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
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			PaddingLeft(1)
)

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
	if len(keys) == 0 {
		return style.Render(" No keys found.")
	}

	listItems := make([]string, max(len(keys), height))
	for i := range listItems {
		if i < len(keys) {
			listItems[i] = keys[i]
		} else {
			listItems[i] = "" // Fill remaining space with empty strings
		}
	}

	l := list.New(listItems).
		ItemStyle(style).
		Enumerator(func(items list.Items, i int) string {
			if i == cur {
				return "▶ " // Current item indicator
			}
			return ""
		}).
		ItemStyleFunc(func(items list.Items, i int) lipgloss.Style {
			if i == cur {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("30")).
					Background(lipgloss.Color("44"))
			}
			return lipgloss.NewStyle()
		})

	return style.Render(l.String())
}
