package list

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
)

var (
	itemStyle         = lipgloss.NewStyle().MarginLeft(2).Foreground(color.White)
	selectedItemStyle = itemStyle.Foreground(color.Primary)
)

type CustomKeyList struct {
	list.Model
}

type item string

func (i item) String() string      { return string(i) }
func (i item) Title() string       { return i.String() }
func (i item) Description() string { return i.Title() }
func (i item) FilterValue() string { return i.Title() }

func New(keys []string, width, height int) CustomKeyList {
	items := make([]list.Item, 0, len(keys))
	for _, k := range keys {
		items = append(items, item(k))
	}
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(color.Primary)
	l := list.New(items, d, width, height)
	l.SetShowTitle(true)
	l.Title = "KEYS"
	l.Styles.Title = l.Styles.Title.Background(color.Primary)
	l.SetShowHelp(false)
	return CustomKeyList{Model: l}
}

func (l *CustomKeyList) Update(msg tea.Msg) (CustomKeyList, tea.Cmd) {
	m, cmd := l.Model.Update(msg)
	return CustomKeyList{Model: m}, cmd
}

func (l *CustomKeyList) View(width, height int) string {
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(color.Primary).Width(width).Height(height)
	l.SetWidth(width - 4)
	l.SetHeight(height)
	return style.Render(l.Model.View())
}
