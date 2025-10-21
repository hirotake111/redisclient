package component

import (
	"log"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/hirotake111/redisclient/internal/color"
)

const (
	hostLabel             = "CONNECTED HOST:"
	dbLabel               = "DATABASE:"
	noKeysFoundMsg        = "No keys found."
	maxHelpMessageHeigtht = 3
)

var (
	// Styles for various UI components
	tabContainerStyle = lipgloss.NewStyle().Padding(0, 1)
	tabLabel          = lipgloss.NewStyle().
				MarginRight(1).
				Background((color.DarkRed)).
				Render(dbLabel)
	tabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(color.Gray)
	activeTabStyle = tabStyle.
			Foreground(color.Red).
			Bold(true).
			Underline(true)
	keyListStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(color.Gray)
	headerStyle      = lipgloss.NewStyle().Padding(0, 1)
	headerLabelStyle = lipgloss.NewStyle().Background(color.DarkRed)
	TitleBarStyle    = lipgloss.NewStyle().PaddingLeft(1)
	filterlabelStyle = lipgloss.NewStyle().PaddingLeft(1).Background(color.DarkRed)
	filterFormStyle  = lipgloss.NewStyle().PaddingLeft(1)

	// help messages
	helpMessages = []string{
		"j or ↓: down",
		"k or ↑: up",
		"Enter: update current value",
		"d: delete key",
		"/: filter keys",
		"n: next page",
		"p: previous page",
		"q/Esc: quit",
	}
	helpTextStyle = lipgloss.NewStyle().
			MarginRight(8).
			Foreground(color.Gray)
)

func Form(label, value string, active bool, width int) string {
	form := lipgloss.NewStyle().
		Width(width / 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(color.Gray)

	if active {
		form = form.BorderForeground(color.Red)
	}

	return form.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		filterlabelStyle.Render(label+":"),
		filterFormStyle.Render(value),
	))

}

func HostHeader(host string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center,
			headerLabelStyle.Render(hostLabel),
			headerStyle.Render(host),
		),
	)
}

func ValueDisplay(value string, width, height int) string {
	maxChrs := (width) * (height) // Adjust for padding and borders
	log.Printf("width  : %d\n", width)
	log.Printf("height : %d\n", height)
	log.Printf("maxChrs: %d\n", maxChrs)
	log.Printf("value : %s\n", value)
	if len(value) > maxChrs {
		value = value[:maxChrs-3] + "..." // Truncate long values
	}
	return lipgloss.NewStyle().
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width).
		Height(height).
		Render(value)
}

func TabRow(tabs int, currentTab int) string {
	_tabs := make([]string, tabs+1)
	_tabs = append(_tabs, tabLabel)
	for i := range tabs {
		if i == currentTab {
			_tabs = append(_tabs, activeTabStyle.Render(strconv.Itoa(i)))
		} else {
			_tabs = append(_tabs, tabStyle.Render(strconv.Itoa(i)))
		}
	}
	return tabContainerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, _tabs...))
}

func TitleBar(title string) lipgloss.Style {
	return TitleBarStyle.SetString(title)
}

func KeyList(keys []string, cur, height, width int, highlighted bool) string {
	style := keyListStyle.Width(width).Height(height)
	if highlighted {
		style = style.BorderForeground(color.Red)
	}
	maxWidthKey := max(0, width-4) // Adjust for padding and borders

	var keyFound = true
	if len(keys) == 0 {
		keys = []string{noKeysFoundMsg}
		keyFound = false
	}

	listItems := make([]string, min(len(keys), height))
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
		Enumerator(func(items list.Items, i int) string {
			if i == cur && keyFound {
				return "▶ " // Current item indicator
			}
			return ""
		}).
		ItemStyleFunc(func(items list.Items, i int) lipgloss.Style {
			if i == cur && keyFound {
				return lipgloss.NewStyle().Background(color.SoftRed)
			}
			return lipgloss.NewStyle()
		})

	return style.Render(l.String())
}

func TTLIndicator(ttl int64) string {
	if ttl < 0 {
		return ""
	}
	if ttl == 0 {
		return ""
	}
	return " (expires in " + strconv.FormatInt(ttl, 10) + " seconds)"
}

func ErrorBox(msg string, width, height int) string {
	var c = color.Gray
	if msg != "" {
		c = color.Red
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(c).
		Foreground(c).
		Render(msg)
}

func HelpPane() string {
	if maxHelpMessageHeigtht == 0 {
		return ""
	}

	t := make([][]string, 0)
	log.Printf("initializing table: %v", t)
	for i, msg := range helpMessages {
		idx := i / maxHelpMessageHeigtht
		if len(t) <= idx {
			t = append(t, make([]string, 0))
		}
		t[idx] = append(t[idx], msg)

	}
	table := make([]string, 0)
	for _, col := range t {
		log.Printf("column: %+v\n", col)
		s := helpTextStyle.Render(lipgloss.JoinVertical(lipgloss.Left, col...))
		table = append(table, s)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, table...)
}
