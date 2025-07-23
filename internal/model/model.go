package model

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/redis/go-redis/v9"
)

const (
	tl  = "╭" // Top left corner for key list
	tr  = "╮" // Top right corner for key list
	bl  = "╰" // Bottom left corner for key list
	br  = "╯" // Bottom right corner for key list
	hl  = "─" // Horizontal line for key list
	vl  = "│" // Vertical line for key list
	dhl = "═" // Double horizontal line for key list
	dvl = "║" // Double vertical line for key list
	tld = "╔" // Top left double corner for key list
	trd = "╗" // Top right double corner for key list
	bld = "╚" // Bottom left double corner for key list
	brd = "╝" // Bottom right double corner for key list
)

var (
	gray  = lipgloss.Color("240") // Gray color for general text
	red   = lipgloss.Color("196") // Red color for error messages
	pink  = lipgloss.Color("205") // Red color for error messages
	green = lipgloss.Color("34")  // Green color for success messages
	blue  = lipgloss.Color("33")  // Blue color for info messages
	white = lipgloss.Color("255") // White color for text

	// Styles for various UI components
	tabStyle = lipgloss.NewStyle().
			Padding(1, 1, 1, 1).
			Foreground(gray)
	activeTabStyle = tabStyle.
			Foreground(pink).
			Bold(true).
			Underline(true)
	keyListStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(gray)
	footerStyle = lipgloss.NewStyle().
			Padding(0, 1)
	footerLabelStyle = lipgloss.NewStyle().
				Background(gray).
				Foreground(white)
)

type State string

const (
	tabSize         = 16 // Default number of database indexes
	ListState State = "list"
)

type Model struct {
	ctx context.Context // Context for app

	currentKeyIdx int        // Current key index in the list
	redisCursor   uint64     // Cursor position in the database
	keys          [][]string // History of keys fetched
	keyHistoryIdx int        // Current index in the key history

	tabs       int // Number of tabs
	currentTab int // Current tab index

	width  int // Width of the terminal window
	height int // Height of the terminal window

	state State         // View state
	redis *redis.Client // Redis client instance
	value string        // Stores the value for the current key

	filterHighlighted bool   // Indicates if the filter form is highlighted
	filterValue       string // Stores the value for the filter form

	displayHelp bool // Flag to display help window
}

func NewModel(ctx context.Context, redis *redis.Client) Model {
	return Model{
		ctx:               ctx,
		tabs:              tabSize,
		currentTab:        0,
		state:             ListState,
		redis:             redis,
		width:             80, // Default width
		height:            24, // Default height
		currentKeyIdx:     0,
		keys:              [][]string{},
		keyHistoryIdx:     0,
		filterHighlighted: false,
		filterValue:       "",
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
	m.redisCursor = msg.RedisCursor
	m.keys = append(m.keys, msg.Keys)
	m.keyHistoryIdx = len(m.keys) - 1 // Reset to the latest history
	m = m.ClearCurrentKeyIdx()
	return m
}

func (m Model) CurrentKeyList() []string {
	if len(m.keys) == 0 {
		log.Print("No keys available in the current history")
		return []string{}
	}
	return m.keys[m.keyHistoryIdx]
}

func (m Model) CurrentKey() string {
	return m.CurrentKeyList()[m.currentKeyIdx]
}

func (m Model) UpdateValue(msg cmd.ValueMsg) Model {
	m.value = string(msg)
	log.Printf("new value: %s", m.value)
	return m
}

func (m Model) NextTab() Model {
	m.currentTab = (m.currentTab + 1) % m.tabs
	return m
}

func (m Model) NextHistory() Model {
	if m.keyHistoryIdx < len(m.keys)-1 {
		m.keyHistoryIdx++
		m = m.ClearCurrentKeyIdx()
		log.Printf("Moved to next history[%d]", m.keyHistoryIdx)
		log.Printf("Current key list: %v", m.CurrentKeyList())
	} else {
		log.Print("No more history to navigate")
	}
	return m
}

func (m Model) PreviousHistory() Model {
	if m.keyHistoryIdx > 0 {
		m.keyHistoryIdx--
		m = m.ClearCurrentKeyIdx()
		log.Printf("Moved to previous history[%d]", m.keyHistoryIdx)
	} else {
		log.Print("No previous history to navigate")
	}
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

func (m Model) MoveCursorDown() Model {
	m.currentKeyIdx = min(m.currentKeyIdx+1, len(m.CurrentKeyList())-1)
	log.Printf("Cursor moved down to index %d", m.currentKeyIdx)
	return m
}

func (m Model) MoveCursorUp() Model {
	m.currentKeyIdx = max(m.currentKeyIdx-1, 0)
	return m
}

func (m Model) ToggleFilterHighlight() Model {
	m.filterHighlighted = !m.filterHighlighted
	log.Printf("Filter form highlight: %t", m.filterHighlighted)
	return m
}

func (m Model) HasNextHistory() bool {
	return m.keyHistoryIdx < len(m.keys)-1
}

func (m Model) HasMoreKeysOnServer() bool {
	return m.redisCursor > 0
}

func (m Model) HasPreviousKeys() bool {
	return m.keyHistoryIdx > 0
}

func (m Model) appendCharToFilterValue(key string) Model {
	m.filterValue += key
	log.Printf("Current filter value: %s", m.filterValue)
	return m
}

func (m Model) removeCharFromFilterValue() Model {
	if len(m.filterValue) > 0 {
		m.filterValue = m.filterValue[:len(m.filterValue)-1]
	}
	log.Printf("Current filter value: %s", m.filterValue)
	return m
}

func (m Model) ClarFilterValue() Model {
	m.filterValue = ""
	log.Print("Clearing filter value")
	return m
}

func (m Model) currentKey() string {
	if m.currentKeyIdx < 0 || m.currentKeyIdx >= len(m.CurrentKeyList()) {
		return ""
	}
	return m.CurrentKeyList()[m.currentKeyIdx]
}

func (m Model) UpdateRedisClient(msg cmd.NewRedisClientMsg) Model {
	m.redis = msg.Redis
	log.Printf("Updating Redis client to %s", m.ConnectionString())
	return m
}

func (m Model) ClearCurrentKeyIdx() Model {
	m.currentKeyIdx = 0
	log.Print("Clearing key index position")
	return m
}

func (m Model) ClearKeyHistory() Model {
	m.keys = [][]string{}
	m.keyHistoryIdx = 0
	log.Print("Clearing key history")
	return m
}

func (m Model) ClearRedisCursor() Model {
	m.redisCursor = 0
	log.Print("Clearing Redis cursor")
	return m
}

func (m Model) DeleteKeyFromList(key string) Model {
	log.Printf("Deleting key %s from current key list", key)
	if len(m.keys) == 0 {
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
	m.keys[m.keyHistoryIdx] = keys
	return m
}

func (m Model) ToggleHelpWindow() Model {
	log.Print("Toggling help window")
	m.displayHelp = !m.displayHelp
	return m
}
