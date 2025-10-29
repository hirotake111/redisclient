package model

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/color"
	"github.com/hirotake111/redisclient/internal/command"
	"github.com/hirotake111/redisclient/internal/component/list"
	"github.com/hirotake111/redisclient/internal/mode"
	"github.com/hirotake111/redisclient/internal/state"
	"github.com/hirotake111/redisclient/internal/util"
	"github.com/hirotake111/redisclient/internal/values"
	"github.com/redis/go-redis/v9"
)

const (
	defaultTabSize = 16 // Default number of database indexes
)

type Model struct {
	ctx    context.Context // Context for app
	width  int             // Width of the terminal window
	height int             // Height of the terminal window
	redis  *redis.Client   // Redis client instance
	mode   *mode.ListMode  // Application mode/state (moved fields)
	State  state.AppState  // Application state
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
		State: state.NewAppState(),
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

func (m Model) UpdateKeyList(msg command.KeysUpdatedMsg) Model {
	m.mode.KeyList = list.New(msg.Keys, 30, 20)
	return m.ReplaceKeys(msg.Keys)
}

func (m Model) UpdateValue(msg command.ValueUpdatedMsg) Model {
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

func (m Model) UpdateRedisClient(msg command.NewRedisClientMsg) Model {
	m.redis = msg.Redis
	log.Printf("Updating Redis client to %s", m.ConnectionString())
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
