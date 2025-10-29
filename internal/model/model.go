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
	"github.com/hirotake111/redisclient/internal/component/viewport"
	"github.com/hirotake111/redisclient/internal/state"
	"github.com/redis/go-redis/v9"
)

const (
	defaultTabSize        = 16 // Default number of database indexes
	defaultKeyListWIdth   = 30
	defaultKeyListHeight  = 20
	defaultViewportWidth  = 50
	defaultViewportHeight = 20
)

type Model struct {
	ctx        context.Context // Context for app
	width      int             // Width of the terminal window
	height     int             // Height of the terminal window
	redis      *redis.Client   // Redis client instance
	State      state.AppState  // Application state
	errorMsg   string
	tabs       int
	currentTab int // Also an index for Redis database
	keyList    list.CustomKeyList
	viewport   viewport.Viewport
}

func NewModel(ctx context.Context, redis *redis.Client) Model {
	return Model{
		ctx:        ctx,
		redis:      redis,
		width:      80,             // Default width
		height:     24,             // Default height
		errorMsg:   "",             // ErrorMsg
		tabs:       defaultTabSize, // Tabs
		currentTab: 0,              // CurrentTab
		keyList:    list.New([]string{}, defaultKeyListWIdth, defaultKeyListHeight),
		viewport:   viewport.New(defaultViewportWidth, defaultViewportHeight),
		State:      state.NewAppState(),
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

func (m Model) NextTab() Model {
	m.currentTab = (m.currentTab + 1) % m.tabs
	return m
}

func (m Model) PreviousTab() Model {
	if m.currentTab == 0 {
		m.currentTab = m.tabs - 1 // Wrap around to the last tab
	} else {
		m.currentTab--
	}
	return m
}

func (m Model) UpdateRedisClient(msg command.NewRedisClientMsg) Model {
	m.redis = msg.Redis
	log.Printf("Updating Redis client to %s", m.ConnectionString())
	return m
}

func (m Model) UpdateErrorMessage(err error) Model {
	log.Println("Updating error message")
	m.errorMsg = err.Error()
	return m
}

func (m Model) ClearErrorMessage() Model {
	log.Print("Clearing error message")
	m.errorMsg = ""
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
