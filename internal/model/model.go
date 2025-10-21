package model

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/mode"
	"github.com/hirotake111/redisclient/internal/values"
	"github.com/redis/go-redis/v9"
)

var (
	// Styles for various UI components
	tabStyle = lipgloss.NewStyle().
			Padding(1, 1, 1, 1).
			Foreground(color.Gray)
	activeTabStyle = tabStyle.
			Foreground(color.DarkRed).
			Bold(true).
			Underline(true)
	keyListStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(color.Gray)
	footerStyle = lipgloss.NewStyle().
			Padding(0, 1)
	footerLabelStyle = lipgloss.NewStyle().
				Background(color.Gray).
				Foreground(color.DarkRed)
)

type State string

const (
	tabSize         = 16 // Default number of database indexes
	ListState State = "list"
)

type Model struct {
	ctx    context.Context // Context for app
	width  int             // Width of the terminal window
	height int             // Height of the terminal window
	redis  *redis.Client   // Redis client instance
	mode   *mode.ListMode  // Application mode/state (moved fields)
	State  State           // Application state (moved from ListMode)
}

func NewModel(ctx context.Context, redis *redis.Client) Model {
	ff := newCustomForm("FILTER: ", "Filter keys...")
	uf := newCustomForm("NEW VALUE: ", "Enter new value...")

	return Model{
		ctx:    ctx,
		redis:  redis,
		width:  80, // Default width
		height: 24, // Default height
		mode: mode.NewListMode(
			"",             // ErrorMsg
			0,              // CurrentKeyIdx
			0,              // RedisCursor
			[][]string{},   // Keys
			0,              // KeyHistoryIdx
			uf,             // UpdateForm
			ff,             // FilterForm
			tabSize,        // Tabs
			0,              // CurrentTab
			values.Value{}, // Value
		),
		State: ListState,
	}
}

func (m Model) HostName() string {
	return m.redis.Options().Addr
}

func (m Model) DB() string {
	return fmt.Sprintf("%d", m.redis.Options().DB)
}
func (m Model) ConnectionString() string {
	return fmt.Sprintf("redis://%s/%d", m.HostName(), m.redis.Options().DB)
}

func (m Model) UpdateWindowSize(height, width int) Model {
	m.width = width
	m.height = height
	return m
}

func (m Model) UpdateKeyList(msg cmd.KeysUpdatedMsg) Model {
	m.mode.RedisCursor = msg.RedisCursor
	m.mode.Keys = append(m.mode.Keys, msg.Keys)
	m.mode.KeyHistoryIdx = len(m.mode.Keys) - 1 // Reset to the latest history
	m = m.ClearCurrentKeyIdx()
	return m
}

func (m Model) CurrentKeyList() []string {
	if len(m.mode.Keys) == 0 {
		log.Print("No keys available in the current history")
		return []string{}
	}
	return m.mode.Keys[m.mode.KeyHistoryIdx]
}

func (m Model) CurrentKey() string {
	return m.CurrentKeyList()[m.mode.CurrentKeyIdx]
}

func (m Model) UpdateValue(msg cmd.ValueUpdatedMsg) Model {
	m.mode.Value = values.NewValue(msg.NewValue, msg.TTL)
	return m
}

func (m Model) NextTab() Model {
	m.mode.CurrentTab = (m.mode.CurrentTab + 1) % m.mode.Tabs
	return m
}

func (m Model) NextHistory() Model {
	if m.mode.KeyHistoryIdx < len(m.mode.Keys)-1 {
		m.mode.KeyHistoryIdx++
		m = m.ClearCurrentKeyIdx()
		log.Printf("Moved to next history[%d]", m.mode.KeyHistoryIdx)
		log.Printf("Current key list: %v", m.CurrentKeyList())
	} else {
		log.Print("No more history to navigate")
	}
	return m
}

func (m Model) PreviousHistory() Model {
	if m.mode.KeyHistoryIdx > 0 {
		m.mode.KeyHistoryIdx--
		m = m.ClearCurrentKeyIdx()
		log.Printf("Moved to previous history[%d]", m.mode.KeyHistoryIdx)
	} else {
		log.Print("No previous history to navigate")
	}
	return m
}

func (m Model) PreviousTab() Model {
	if m.mode.CurrentTab == 0 {
		m.mode.CurrentTab = m.mode.Tabs - 1 // Wrap around to the last tab
	} else {
		m.mode.CurrentTab--
	}
	return m
}

func (m Model) MoveCursorDown() Model {
	m.mode.CurrentKeyIdx = min(m.mode.CurrentKeyIdx+1, len(m.CurrentKeyList())-1)
	log.Printf("Cursor moved down to index %d", m.mode.CurrentKeyIdx)
	return m
}

func (m Model) MoveCursorUp() Model {
	m.mode.CurrentKeyIdx = max(m.mode.CurrentKeyIdx-1, 0)
	return m
}

func (m Model) HasNextHistory() bool {
	return m.mode.KeyHistoryIdx < len(m.mode.Keys)-1
}

func (m Model) HasMoreKeysOnServer() bool {
	return m.mode.RedisCursor > 0
}

func (m Model) HasPreviousKeys() bool {
	return m.mode.KeyHistoryIdx > 0
}

func (m Model) currentKey() string {
	if m.mode.CurrentKeyIdx < 0 || m.mode.CurrentKeyIdx >= len(m.CurrentKeyList()) {
		return ""
	}
	return m.CurrentKeyList()[m.mode.CurrentKeyIdx]
}

func (m Model) UpdateRedisClient(msg cmd.NewRedisClientMsg) Model {
	m.redis = msg.Redis
	log.Printf("Updating Redis client to %s", m.ConnectionString())
	return m
}

func (m Model) ClearCurrentKeyIdx() Model {
	m.mode.CurrentKeyIdx = 0
	log.Print("Clearing key index position")
	return m
}

func (m Model) ClearKeyHistory() Model {
	m.mode.Keys = [][]string{}
	m.mode.KeyHistoryIdx = 0
	log.Print("Clearing key history")
	return m
}

func (m Model) ClearRedisCursor() Model {
	m.mode.RedisCursor = 0
	log.Print("Clearing Redis cursor")
	return m
}

func (m Model) DeleteKeyFromList(key string) Model {
	log.Printf("Deleting key %s from current key list", key)
	if len(m.mode.Keys) == 0 {
		log.Print("No keys available to delete")
		return m
	}

	if key == "" {
		log.Print("Empty key provided for deletion - ignoring")
		return m
	}
	keys := make([]string, 0, len(m.CurrentKeyList())-1)
	for _, k := range m.CurrentKeyList() {
		if k != key {
			keys = append(keys, k)
		}
	}
	m.mode.Keys[m.mode.KeyHistoryIdx] = keys
	return m
}

func (m Model) ToListState() Model {
	log.Print("Switching to list state")
	m.State = ListState
	return m
}

func (m Model) EmptyValue() Model {
	log.Print("Clearing value")
	m.mode.Value = values.Value{}
	return m
}

func (m Model) UpdateErrorMessage(err error) Model {
	log.Println("Updating error message")
	m.mode.ErrorMsg = err.Error()
	return m
}

func (m Model) ClearErrorMessage() Model {
	log.Print("Clearing error message")
	m.mode.ErrorMsg = ""
	return m
}

func newCustomForm(prompt, placeholder string) *textarea.Model {
	ff := textarea.New()
	ff.Prompt = prompt
	ff.Placeholder = placeholder
	ff.SetHeight(1)
	ff.SetWidth(100)
	ff.CharLimit = 100
	ff.KeyMap.InsertNewline.SetEnabled(false) // Disable newline insertion
	ff.BlurredStyle.Base = ff.BlurredStyle.Base.Border(lipgloss.RoundedBorder()).BorderForeground(color.Gray)
	ff.FocusedStyle.Base = ff.FocusedStyle.Base.Border(lipgloss.RoundedBorder()).BorderForeground(color.DarkRed)
	ff.ShowLineNumbers = false
	ff.Blur()
	return &ff
}
