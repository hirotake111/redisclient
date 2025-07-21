package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/hirotake111/redisclient/internal/config"
	"github.com/hirotake111/redisclient/internal/keylist"
	"github.com/hirotake111/redisclient/internal/logger"
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
	ListState State = "list"
)

type model struct {
	ctx context.Context // Context for app

	keys          []string // List of keys fetched from redis
	currentKeyIdx int      // Current key index in the list

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

func (m model) HostName() string {
	return m.redis.Options().Addr
}

func (m model) DB() string {
	return fmt.Sprintf("%d", m.redis.Options().DB)
}
func (m model) ConnectionString() string {
	return fmt.Sprintf("redis://%s/%d", m.HostName(), m.redis.Options().DB)
}

func (m model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case "list":
		if m.filterHighlighted {
			// filter mode activated
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
				// Handle filter input
				m = m.appendCharToFilterValue(key)
				return m, nil
			}
		}

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
				return m, cmd.GetValue(m.ctx, m.redis, []string{m.currentKey()})
			}
			if key == "k" {
				log.Print("Moving cursor up")
				m = m.MoveCursorUp()
				return m, cmd.GetValue(m.ctx, m.redis, []string{m.currentKey()})
			}
			if key == "/" {
				m = m.ToggleFilterHighlight()
				return m, nil
			}
			if key == tea.KeyTab.String() {
				m = m.NextTab()
				return m, cmd.UpdateDatabase(m.ctx, m.redis, m.currentTab)
			}
			if key == tea.KeyShiftTab.String() {
				m = m.PreviousTab()
				return m, cmd.UpdateDatabase(m.ctx, m.redis, m.currentTab)
			}

		case cmd.ValueMsg:
			return m.UpdateValue(msg), nil

		case cmd.KeysUpdatedMsg:
			return m.UpdateKeyList(msg), cmd.GetValue(m.ctx, m.redis, msg) // Fetch value for the first key
		case cmd.NewRedisClientMsg:
			log.Print("Received new Redis client message")
			m = m.UpdateRedisClient(msg)
			return m, cmd.GetKeys(m.ctx, m.redis)
		}

	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case ListState:
		return keylist.Render(
			m.width,
			m.height,
			m.tabs,
			m.currentTab,
			m.keys,
			m.currentKeyIdx,
			m.value,
			m.HostName(),
		)
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
}

func (m model) UpdateWindowSize(height, width int) model {
	m.width = width
	m.height = height
	return m
}

func (m model) UpdateKeyList(msg cmd.KeysUpdatedMsg) model {
	m.keys = msg
	return m
}

func (m model) UpdateValue(msg cmd.ValueMsg) model {
	m.value = string(msg)
	log.Printf("new value: %s", m.value)
	return m
}

func (m model) NextTab() model {
	m.currentTab = (m.currentTab + 1) % m.tabs
	return m
}

func (m model) PreviousTab() model {
	if m.currentTab == 0 {
		m.currentTab = m.tabs - 1 // Wrap around to the last tab
	} else {
		m.currentTab--
	}
	return m
}

func (m model) MoveCursorDown() model {
	m.currentKeyIdx = min(m.currentKeyIdx+1, len(m.keys)-1)
	return m
}

func (m model) MoveCursorUp() model {
	m.currentKeyIdx = max(m.currentKeyIdx-1, 0)
	return m
}

func (m model) ToggleFilterHighlight() model {
	m.filterHighlighted = !m.filterHighlighted
	log.Printf("Filter form highlight: %t", m.filterHighlighted)
	return m
}

func (m model) appendCharToFilterValue(key string) model {
	m.filterValue += key
	log.Printf("Current filter value: %s", m.filterValue)
	return m
}

func (m model) removeCharFromFilterValue() model {
	if len(m.filterValue) > 0 {
		m.filterValue = m.filterValue[:len(m.filterValue)-1]
	}
	log.Printf("Current filter value: %s", m.filterValue)
	return m
}

func (m model) currentKey() string {
	if m.currentKeyIdx < 0 || m.currentKeyIdx >= len(m.keys) {
		return ""
	}
	return m.keys[m.currentKeyIdx]
}

func (m model) UpdateRedisClient(msg cmd.NewRedisClientMsg) model {
	m.redis = msg.Redis
	log.Printf("Updating Redis client to %s", m.ConnectionString())
	return m
}

func main() {
	// Initialize logger to write to temp file
	if err := logger.Initialize(); err != nil {
		// If logger fails, print to stderr and exit
		log.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Print("Logger initialized successfully")

	ctx := context.Background()

	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		log.Printf("Failed to get config from environment: %v\n", err)
		os.Exit(1)
	}

	r := redis.NewClient(cfg.Option)
	if _, err := r.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s - %v", cfg.Option.Addr, err)
		os.Exit(1)
	}

	m := model{
		ctx:   ctx,
		tabs:  16,
		state: "list",
		redis: r,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	log.Print("Program exited successfully")
}
