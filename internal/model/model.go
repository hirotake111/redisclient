package model

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/apperror"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/mode"
	"github.com/hirotake111/redisclient/internal/util"
	"github.com/hirotake111/redisclient/internal/values"
	"github.com/redis/go-redis/v9"
)

var (
	// Styles for various UI components
	tabStyle = lipgloss.NewStyle().
		Padding(1, 1, 1, 1).
		Foreground(color.Primary)
	// activeTabStyle = tabStyle.
	// 		Foreground(color.Secondary).
	// 		Bold(true).
	// 		Underline(true)
	// keyListStyle = lipgloss.NewStyle().
	// 		BorderStyle(lipgloss.RoundedBorder()).
	// 		BorderForeground(color.Primary)
	// footerStyle = lipgloss.NewStyle().
	// 		Padding(0, 1)
	// footerLabelStyle = lipgloss.NewStyle().
	// 			Background(color.Primary).
	// 			Foreground(color.Secondary)
)

type State string

const (
	defaultTabSize       = 16 // Default number of database indexes
	ListState      State = "list"
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
			[]string{},     // Keys
			uf,             // UpdateForm
			ff,             // FilterForm
			defaultTabSize, // Tabs
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
	return m.ReplaceKeys(msg.Keys)
}

func (m Model) CurrentKey() string {
	return m.mode.Keys[m.mode.CurrentKeyIdx]
}

func (m Model) UpdateValue(msg cmd.ValueUpdatedMsg) Model {
	m.mode.Value = values.NewValue(msg.NewValue, msg.TTL)
	return m
}

func (m Model) NextTab() Model {
	m.mode.CurrentTab = (m.mode.CurrentTab + 1) % m.mode.Tabs
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

func (m Model) MoveCursorDown() (Model, error) {
	if len(m.mode.Keys) == 0 {
		return m, apperror.CantMoveCursorDownError
	}
	if m.mode.CurrentKeyIdx+1 == len(m.mode.Keys) {
		return m, apperror.CantMoveCursorDownError
	}
	m.mode.CurrentKeyIdx++
	log.Printf("Cursor moved down to index %d", m.mode.CurrentKeyIdx)
	return m, nil
}

func (m Model) MoveCursorUp() (Model, error) {
	if len(m.mode.Keys) == 0 {
		return m, apperror.CantMoveCursorUpError
	}
	if m.mode.CurrentKeyIdx-1 < 0 {
		return m, apperror.CantMoveCursorUpError
	}
	m.mode.CurrentKeyIdx--
	log.Printf("Cursor moved up to index %d", m.mode.CurrentKeyIdx)
	return m, nil
}

func (m Model) currentKey() string {
	if m.mode.CurrentKeyIdx < 0 || m.mode.CurrentKeyIdx >= len(m.mode.Keys) {
		return ""
	}
	return m.mode.Keys[m.mode.CurrentKeyIdx]
}

func (m Model) UpdateRedisClient(msg cmd.NewRedisClientMsg) Model {
	m.redis = msg.Redis
	log.Printf("Updating Redis client to %s", m.ConnectionString())
	return m
}

func (m Model) ResetKeyIndex() Model {
	m.mode.CurrentKeyIdx = 0
	log.Print("Reset key index position to 0")
	return m
}

func (m Model) DeleteKeyFromList(key string) Model {
	log.Printf("Deleting key %s from current key list", key)
	if len(m.mode.Keys) == 0 {
		log.Print("No keys available to delete - skip deletion")
		return m
	}

	if key == "" {
		log.Print("Empty key provided for deletion - ignoring")
		return m
	}

	m.mode.Keys = util.Filter(m.mode.Keys, func(k string) bool { return k != key })
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

func (m Model) ReplaceKeys(keys []string) Model {
	log.Printf("Replacing key list with %d new keys", len(keys))
	m.mode.Keys = keys
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
	ff.BlurredStyle.Base = ff.BlurredStyle.Base.Border(lipgloss.RoundedBorder()).BorderForeground(color.Grey)
	ff.FocusedStyle.Base = ff.FocusedStyle.Base.Border(lipgloss.RoundedBorder()).BorderForeground(color.Primary)
	ff.ShowLineNumbers = false
	ff.Blur()
	return &ff
}
