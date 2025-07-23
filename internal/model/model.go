package model

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/hirotake111/redisclient/internal/keylist"
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

func (m Model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis, m.redisCursor)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case "list":
		if m.filterHighlighted {
			//
			// FILTER MODE ACTIVATED
			//
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				return m.UpdateWindowSize(msg.Height, msg.Width), nil
			case tea.KeyMsg:
				key := msg.String()
				if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() {
					log.Print("Exiting filter mode")
					m = m.ToggleFilterHighlight()
					return m, nil
				}
				if key == tea.KeyBackspace.String() {
					m = m.removeCharFromFilterValue()
					return m, nil
				}
				if key == tea.KeyEnter.String() {
					m = m.ToggleFilterHighlight()
					log.Printf("Filter applied: %s", m.filterValue)
					return m, nil
				}
				// Handle filter input
				m = m.appendCharToFilterValue(key)
				return m, nil
			}
		}

		// filter mode not activated
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			return m.UpdateWindowSize(msg.Height, msg.Width), nil
		case tea.KeyMsg:
			key := msg.String()
			if key == tea.KeyEsc.String() || key == tea.KeyCtrlC.String() || key == "q" {
				return m, tea.Quit
			}
			if key == "j" {
				log.Print("Moving cursor down")
				m = m.MoveCursorDown()
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey())
			}
			if key == "k" {
				log.Print("Moving cursor up")
				m = m.MoveCursorUp()
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey())
			}
			if key == "/" {
				log.Print("Filter mode activated")
				m = m.ToggleFilterHighlight()
				m = m.ClarFilterValue()
				return m, nil
			}
			if key == "n" {
				log.Print("key 'n' pressed, moving to next key list")
				if m.HasNextHistory() {
					log.Print("Next history exists, moving to next key list")
					m = m.NextHistory()
					return m, cmd.GetValue(m.ctx, m.redis, m.currentKey()) // Fetch value for the current key
				}
				if m.HasMoreKeysOnServer() {
					log.Print("Fetching a next key list from server")
					return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor) // Fetch keys for the new tab
				}
				log.Print("No more keys to fetch")
				return m, nil
			}
			if key == "p" {
				log.Print("key 'p' pressed, moving to previous key list")
				if m.HasPreviousKeys() {
					log.Print("Moving to previous key list")
					m = m.PreviousHistory()
				} else {
					log.Print("No previous keys to fetch")
				}
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey()) // Fetch value for the current key
			}
			if key == tea.KeyTab.String() {
				m = m.NextTab()
				return m, cmd.UpdateDatabase(m.ctx, m.redis, m.currentTab)
			}
			if key == tea.KeyShiftTab.String() {
				m = m.PreviousTab()
				return m, cmd.UpdateDatabase(m.ctx, m.redis, m.currentTab)
			}
			if key == "d" {
				log.Print("key 'd' pressed, deleting current key")
				currentKey := m.currentKey()
				if currentKey == "" {
					log.Print("No current key selected for deletion")
					return m, nil
				}
				log.Printf("Deleting key: %s", currentKey)
				return m, cmd.DeleteKey(m.ctx, m.redis, currentKey)
			}

		case cmd.ValueMsg:
			return m.UpdateValue(msg), nil

		case cmd.KeysUpdatedMsg:
			log.Printf("Received keys updated message. len: %d. cursor: %d", len(msg.Keys), msg.RedisCursor)
			m = m.UpdateKeyList(msg)
			if len(msg.Keys) == 0 {
				log.Print("No keys found, returning empty value")
				m.value = ""
				return m, cmd.DisplayEmptyValue
			}
			return m, cmd.GetValue(m.ctx, m.redis, m.currentKey()) // Fetch value for the first key

		case cmd.KeyDeletedMsg:
			log.Printf("Received key deleted message for key: %s", msg.Key)
			m = m.DeleteKeyFromList(msg.Key)
			return m, nil

		case cmd.NewRedisClientMsg:
			log.Print("Received new Redis client message")
			m = m.UpdateRedisClient(msg).ClearCurrentKeyIdx().ClearKeyHistory().ClearRedisCursor()
			return m, cmd.GetKeys(m.ctx, m.redis, m.redisCursor)

		case cmd.ErrMsg:
			log.Printf("Error occurred: %s", msg.Err)
		}

	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case ListState:
		return keylist.Render(
			m.width,
			m.height,
			m.tabs,
			m.currentTab,
			m.CurrentKeyList(),
			m.currentKeyIdx,
			m.value,
			m.HostName(),
			m.filterHighlighted,
			m.filterValue,
		)
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
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
