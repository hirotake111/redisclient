package list

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/hirotake111/redisclient/internal/color"
)

var (
	itemStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(color.White)
)

type CustomKeyList struct {
	list.Model
	keyIndex map[string]int
}

type item string

func (i item) String() string      { return string(i) }
func (i item) Title() string       { return i.String() }
func (i item) Description() string { return i.Title() }
func (i item) FilterValue() string { return i.Title() }

func New(keys []string, width, height int) CustomKeyList {
	items := make([]list.Item, 0, len(keys))
	ki := make(map[string]int, len(keys))
	for i, k := range keys {
		items = append(items, item(k))
		ki[k] = i
	}
	// log.Printf("Key index map: %+v\n", ki)
	return CustomKeyList{
		Model:    newList(items, width, height),
		keyIndex: ki,
	}
}

func newList(items []list.Item, widt, height int) list.Model {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(color.Primary)
	l := list.New(items, d, widt, height)
	l.SetShowTitle(true)
	l.Title = "KEYS"
	l.Styles.Title = l.Styles.Title.Background(color.Primary)
	l.SetShowHelp(false)
	return l
}

func (l *CustomKeyList) Update(msg tea.Msg) (CustomKeyList, tea.Cmd) {
	if msg, ok := msg.(cmd.KeyDeletedMsg); ok {
		log.Printf("DEBUG: %+v\n", l.keyIndex)
		items := l.Model.Items()
		if idx, ok := l.keyIndex[msg.Key]; ok {
			log.Printf("Item to be deleted: %d - %s\n", idx, msg.Key)
			delete(l.keyIndex, msg.Key)
			items = append(items[:idx], items[idx+1:]...)
		} else {
			log.Printf("Key to be deleted not found in index map: %s\n", msg.Key)
		}
		return CustomKeyList{
			Model:    newList(items, l.Model.Width(), l.Model.Height()),
			keyIndex: l.keyIndex,
		}, nil
	}

	m, cmd := l.Model.Update(msg)
	return CustomKeyList{
		Model:    m,
		keyIndex: l.keyIndex,
	}, cmd
}

func (l *CustomKeyList) View(width, height int) string {
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(color.Primary).Width(width).Height(height)
	l.SetWidth(width - 4)
	l.SetHeight(height)
	return style.Render(l.Model.View())
}
