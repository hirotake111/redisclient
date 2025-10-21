package helpbox

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
)

var (
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

func New(height int) string {

	t := make([][]string, 0)
	for i, msg := range helpMessages {
		idx := i / height
		if len(t) <= idx {
			t = append(t, make([]string, 0))
		}
		t[idx] = append(t[idx], msg)

	}
	table := make([]string, 0)
	for _, col := range t {
		s := helpTextStyle.Render(lipgloss.JoinVertical(lipgloss.Left, col...))
		table = append(table, s)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, table...)
}
