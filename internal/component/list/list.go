package list

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/domain/infoid"
	"github.com/hirotake111/redisclient/internal/state"
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
	model list.Model
}

type item string

func (i item) String() string      { return string(i) }
func (i item) Title() string       { return i.String() }
func (i item) Description() string { return i.Title() }
func (i item) FilterValue() string { return i.Title() }

func New(keys []string, width, height int) CustomKeyList {
	return CustomKeyList{
		model: newItems(keys, width, height),
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
	if !st.ListActive() {
		return l, nil
	}

	var cmds []tea.Cmd
	prv := empty
	if l.model.SelectedItem() != nil {
		prv = l.model.SelectedItem().FilterValue()
	}

	if msg, ok := msg.(command.KeyDeletedMsg); ok {
		l.removeKeyFromList()
		t := fmt.Sprintf("Key '%s' deleted successfully.", msg.Key)
		cmds = append(cmds, command.NewInfoInfoCmd(infoid.New(), t, 5*time.Second))
	}

	if msg, ok := msg.(command.KeysUpdatedMsg); ok {
		prev := l.model.SelectedItem()
		items := newItems(msg.Keys, l.model.Width(), l.model.Height())
		l.model = items
		// Restore previous cursor position
		if prev != nil {
			pi, ok := prev.(item)
			if !ok {
				cmds = append(cmds, command.NewErrorInfoCmd(infoid.New(), fmt.Errorf("failed to assert previous item type"), 5*time.Second))
			}
			for i, a := range l.model.Items() {
				if a.(item).Title() == pi.Title() {
					log.Printf("Restoring cursor position to index %d for item: %+v", i, a)
					l.model.Select(i)
					break
				}
			}

			if selected := l.model.SelectedItem(); selected != nil {
				cmds = append(cmds, command.GetValue(ctx, client, selected.FilterValue()))
			}
		}

		return l, tea.Batch(cmds...)
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		log.Println("Processing key message in CustomKeyList")
		key := msg.String()
		switch {
		case key == "enter" && l.model.FilterState() != list.Filtering:
			log.Print("key 'enter' pressed, activating viewport")
			cmds = append(cmds, state.ActivateViewportCmd)

		case key == "x":
			l, cmds = l.DeleteKey(ctx, client, cmds)

		case key == "X":
			l, cmds = l.BulkDelete(ctx, client, cmds)

		case key == "r":
			// Avoid refreshing while filtering (otherwise it gets refreshed when pressing r key)
			if l.model.FilterState() != list.Filtering {
				log.Print("key 'r' pressed, refreshing key list")
				cmds = append(cmds, command.GetKeys(ctx, client, ""))
			}

		case key == "y":
			if l.model.FilterState() != list.Filtering {
				log.Print("key 'y' pressed, copying current key to clipboard")
				cmds = append(cmds, command.CopyValueToClipboard(ctx, l.model.SelectedItem().FilterValue()))
			}
		}
	}

	m, cmd := l.model.Update(msg)
	l.model = m
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if l.ShouldUpdateValue(prv) {
		cmds = append(cmds, command.GetValue(ctx, client, l.model.SelectedItem().FilterValue()))
	} else {
		log.Print("No change in selected key")
	}

	log.Printf("End of CustomKeyList.Update - total cmds: %v", cmds)
	return l, tea.Batch(cmds...)
}

func (l *CustomKeyList) View(width, height int, st state.AppState) string {
	l.model.SetWidth(width - 4)
	l.model.SetHeight(height)
	style := defaultContainer
	if st.ListActive() {
		style = activeContainer
	}
	return style.Width(width).Height(height).Render(l.model.View())
}

func (l CustomKeyList) DeleteKey(ctx context.Context, client *redis.Client, cmds []tea.Cmd) (CustomKeyList, []tea.Cmd) {
	log.Print("key 'x' pressed, deleting current key")
	si := l.model.SelectedItem()
	if si == nil {
		log.Print("No item selected - skipping deletion")
		return l, cmds
	}

	k := si.FilterValue()
	if k == "" {
		log.Print("No current key selected for deletion")
		return l, cmds
	}

	log.Printf("Deleting key: %s", k)
	cmds = append(cmds, command.DeleteKey(ctx, client, k))
	return l, cmds
}

func (l CustomKeyList) BulkDelete(ctx context.Context, client *redis.Client, cmds []tea.Cmd) (CustomKeyList, []tea.Cmd) {
	if len(l.model.VisibleItems()) == 0 {
		log.Print("No visible items to delete in bulk - skipping deletion")
		return l, cmds
	}

	log.Printf("key 'X' pressed, perform bulk delete for %d keys", len(l.model.VisibleItems()))
	keys := make([]string, 0, len(l.model.VisibleItems()))
	for _, it := range l.model.VisibleItems() {
		keys = append(keys, it.FilterValue())
	}
	cmds = append(cmds, command.BulkDelete(ctx, client, keys))
	return l, cmds
}
func (l *CustomKeyList) removeKeyFromList() {
	selected := l.model.SelectedItem().FilterValue()
	log.Printf("Removing selected item \"%s\" at index %d. items(length: %d)", selected, l.model.GlobalIndex(), len(l.model.Items()))
	l.model.RemoveItem(l.model.GlobalIndex())
	log.Printf("Removed  selected item \"%s\". items(length: %d)", selected, len(l.model.Items()))
	if l.model.FilterState() == list.FilterApplied {
		// Manually re-apply filter to update visible items
		si := l.model.Index()
		l.model.SetFilterText(l.model.FilterValue())
		l.model.Select(si)
		if len(l.model.VisibleItems()) == 0 {
			log.Println("Clear filter text as no items are visible after deletion")
			l.model.SetFilterState(list.Unfiltered)
		}
	}
}

func (l CustomKeyList) ShouldUpdateValue(prv string) bool {
	log.Printf("%d items after update. Index: %d", len(l.model.Items()), l.model.Index())
	si := l.model.SelectedItem()
	log.Printf("Selected items after update: %+v. Index: %d, global index: %d", si, l.model.Index(), l.model.GlobalIndex())
	if si == nil {
		return false
	}

	log.Printf("Before getting current key. Item: %+v", si)
	cur := si.FilterValue()
	log.Printf("Current selected key: \"%s\", previous selected key: \"%s\"", cur, prv)
	return prv != cur
}

func (l CustomKeyList) IsBeingUnfiltered() bool {
	return l.model.FilterState() == list.Unfiltered
}
func (l CustomKeyList) IsFitering() bool {
	return l.model.FilterState() == list.Filtering
}
