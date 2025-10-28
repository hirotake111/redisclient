package list

import (
	"context"
	"log"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/redis/go-redis/v9"
)

const (
	empty = "(empty)"
)

type arrayElement struct {
	key   string
	index int
}

type CustomKeyList struct {
	list.Model
	sorted []arrayElement
}

type item string

func (i item) String() string      { return string(i) }
func (i item) Title() string       { return i.String() }
func (i item) Description() string { return i.Title() }
func (i item) FilterValue() string { return i.Title() }

func New(keys []string, width, height int) CustomKeyList {
	items := make([]list.Item, 0, len(keys))
	ki := make(map[string]int, len(keys))
	arr := make([]arrayElement, 0, len(keys))
	for i, k := range keys {
		items = append(items, item(k))
		ki[k] = i
		arr = append(arr, arrayElement{key: k, index: i})
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i].key < arr[j].key
	})
	// log.Printf("Key index map: %+v\n", ki)
	return CustomKeyList{
		Model:  newList(items, width, height),
		sorted: arr,
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

func (l *CustomKeyList) Update(ctx context.Context, client *redis.Client, msg tea.Msg) (CustomKeyList, tea.Cmd) {
	var cmds []tea.Cmd
	prv, cur := empty, empty
	if l.SelectedItem() != nil {
		prv = l.Model.SelectedItem().FilterValue()
	}
	log.Printf("CustomKeyList received message: %+v\n", msg)
	log.Printf("Key before update: \"%+v\"", prv)
	if _, ok := msg.(command.KeyDeletedMsg); ok {
		selected := l.Model.SelectedItem().FilterValue()
		log.Printf("Removing selected item \"%s\" at index %d. items(%d): %+v", selected, l.GlobalIndex(), len(l.Model.Items()), l.Model.Items())
		l.Model.RemoveItem(l.GlobalIndex())
		log.Printf("Removed  selected item \"%s\".            items(%d): %+v", selected, len(l.Model.Items()), l.Model.Items())
		if l.FilterState() == list.FilterApplied {
			si := l.Index()
			l.SetFilterText(l.FilterValue())
			l.Select(si)
			if len(l.VisibleItems()) == 0 {
				log.Println("Clear filter text as no items are visible after deletion")
				l.SetFilterState(list.Unfiltered)
			}
		}
	}

	m, cmd := l.Model.Update(msg)
	cmds = append(cmds, cmd)
	log.Printf("Items after update: %+v. Index: %d", m.Items(), m.Index())
	selectedItem := m.SelectedItem()
	log.Printf("Selected items after update: %+v. Index: %d, global index: %d", selectedItem, m.Index(), m.GlobalIndex())
	log.Printf("visible: %+v", m.VisibleItems())
	if selectedItem != nil {
		log.Printf("Before getting current key. Item: %+v", selectedItem)
		cur = selectedItem.FilterValue()
	}
	if cur != empty && prv != cur {
		log.Printf("Will be updating value display for key: \"%s\"", cur)
		cmds = append(cmds, command.GetValue(ctx, client, cur))
	} else {
		log.Printf("No change in selected key: \"%s\", prev: \"%s\"", cur, prv)
	}

	return CustomKeyList{
		Model:  m,
		sorted: l.sorted,
	}, tea.Batch(cmds...)
}

func (l *CustomKeyList) View(width, height int) string {
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(color.Primary).Width(width).Height(height)
	l.SetWidth(width - 4)
	l.SetHeight(height)
	return style.Render(l.Model.View())
}
