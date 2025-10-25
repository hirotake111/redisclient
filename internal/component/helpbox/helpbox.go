package helpbox

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
)

var (
	helpMessages = [][2]string{
		{"j or ↓", "down"},
		{"k or ↑", "up"},
		{"h or ←", "left"},
		{"l or →", "right"},
		{"Enter", "update current value"},
		{"d", "delete key"},
		{"/", "filter keys"},
		{"n", "next page"},
		{"p", "previous page"},
		{"q/string{", " Esc: quit"},
	}
	helpTextkeyStyle = lipgloss.NewStyle().
				MarginRight(1).
				Foreground(color.Grey)
	helpTextValueStyle = helpTextkeyStyle.Foreground(color.Primary)
	columnStyle        = lipgloss.NewStyle().MarginRight(8)
)

func New(height int) string {

	tbl := make([][]string, 0)
	for i, arr := range helpMessages {
		idx := i / height
		if len(tbl) <= idx {
			tbl = append(tbl, make([]string, 0))
		}
		tbl[idx] = append(tbl[idx], tooltipText(arr[0], arr[1]))

	}
	table := make([]string, 0)
	for _, txts := range tbl {
		col := columnStyle.Render(lipgloss.JoinVertical(lipgloss.Left, txts...))
		table = append(table, col)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, table...)
}

func tooltipText(key, val string) string {
	ks := helpTextkeyStyle.Render(key + ":")
	vs := helpTextValueStyle.Render(val)
	return ks + vs
}
