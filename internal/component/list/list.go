package list

import (
	"context"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/state"
	"github.com/hirotake111/redisclient/internal/util"
	"github.com/redis/go-redis/v9"
)

const (
	empty = "(empty)"
)

var (
	defaultContainer = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(color.Primary)
	activeContainer  = defaultContainer.BorderStyle(lipgloss.ThickBorder())
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
	return CustomKeyList{
		Model: newItems(keys, width, height),
	}
}

func newItems(keys []string, widt, height int) list.Model {
	items := make([]list.Item, 0, len(keys))
	for _, k := range keys {
		items = append(items, item(k))
	}
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

func (l CustomKeyList) Update(ctx context.Context, client *redis.Client, msg tea.Msg, st state.AppState) (CustomKeyList, tea.Cmd) {
	util.LogMsg("CustomKeyList received a message", msg)

	if !st.ListActive() {
		return l, nil
	}

	var cmds []tea.Cmd
	prv, cur := empty, empty
	if l.SelectedItem() != nil {
		prv = l.Model.SelectedItem().FilterValue()
	}

	if _, ok := msg.(command.KeyDeletedMsg); ok {
		selected := l.Model.SelectedItem().FilterValue()
		log.Printf("Removing selected item \"%s\" at index %d. items(length: %d)", selected, l.GlobalIndex(), len(l.Model.Items()))
		l.Model.RemoveItem(l.GlobalIndex())
		log.Printf("Removed  selected item \"%s\". items(length: %d)", selected, len(l.Model.Items()))
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

	if msg, ok := msg.(command.KeysUpdatedMsg); ok {
		items := newItems(msg.Keys, l.Width(), l.Height())
		l.Model = items
		selected := l.SelectedItem()
		return l, command.GetValue(ctx, client, selected.FilterValue())
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		key := msg.String()
		if key == "enter" && l.FilterState() != list.Filtering {
			// Send command to activate viewport
			cmds = append(cmds, state.ActivateViewportCmd)
		}

		if key == "x" {
			log.Print("key 'x' pressed, deleting current key")
			currentKey := l.Model.SelectedItem().FilterValue()
			if currentKey == "" {
				log.Print("No current key selected for deletion")
			} else {
				log.Printf("Deleting key: %s", currentKey)
				cmds = append(cmds, command.DeleteKey(ctx, client, currentKey))
			}
		}

		if key == "X" && l.FilterState() == list.FilterApplied {
			log.Printf("key \"X\" pressed, perform bulk delete for %d keys", len(l.VisibleItems()))
			keys := make([]string, 0, len(l.VisibleItems()))
			for _, it := range l.VisibleItems() {
				keys = append(keys, it.FilterValue())
			}
			cmds = append(cmds, command.BulkDelete(ctx, client, keys))
		}

	}

	m, cmd := l.Model.Update(msg)
	cmds = append(cmds, cmd)
	log.Printf("%d items after update. Index: %d", len(m.Items()), m.Index())
	selectedItem := m.SelectedItem()
	log.Printf("Selected items after update: %+v. Index: %d, global index: %d", selectedItem, m.Index(), m.GlobalIndex())
	log.Printf("visible: %d", len(m.VisibleItems()))
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

	l.Model = m
	return l, tea.Batch(cmds...)
}

func (l *CustomKeyList) View(width, height int, st state.AppState) string {
	l.SetWidth(width - 4)
	l.SetHeight(height)
	style := defaultContainer
	if st.ListActive() {
		style = activeContainer
	}
	return style.Width(width).Height(height).Render(l.Model.View())
}
