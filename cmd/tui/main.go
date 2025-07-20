package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/hirotake111/redisclient/internal/cmd"
	"github.com/hirotake111/redisclient/internal/config"
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
	tabStyle       = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("240")).PaddingTop(1).PaddingBottom(1)
	activeTabStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("205")).PaddingTop(1).PaddingBottom(1).Bold(true).Underline(true)
	keyListStyle   = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

type State string

const (
	ListState State = "list"
)

type model struct {
	ctx context.Context // Context for app

	keys          []string // List of keys fetched from redis
	currentKeyIdx int      // Current key index in the list

	tabs       []string // List of tabs
	currentTab int      // Current tab index

	width   int           // Width of the terminal window
	height  int           // Height of the terminal window
	state   State         // View state
	redis   *redis.Client // Redis client instance
	message string        // temporary message for display
	value   string        // Stores the value for the current key
}

func (m model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case "list":
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			return m.UpdateWindowSize(msg.Height, msg.Width), nil
		case tea.KeyMsg:
			key := msg.String()
			if key == "esc" || key == "ctrl+c" || key == "ctrl+q" {
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
			if key == "tab" {
				// Switch to next tabIndex
				m = m.NextTab()
				return m, cmd.GetValue(m.ctx, m.redis, m.currentKey())
			}

		case cmd.ValueMsg:
			return m.UpdateValue(msg), nil

		case cmd.KeysUpdatedMsg:
			if len(msg) > 0 {
				return m.UpdateKeyList(msg), cmd.GetValue(m.ctx, m.redis, msg[0]) // Fetch value for the first key
			}
			return m.UpdateKeyList(msg), nil
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case ListState:
		// Calculate widths
		widthKeyListView := m.width / 3
		widthValueView := m.width - widthKeyListView - 10 // Adjust for padding and borders

		tabRow := renderTabRow(m.tabs, m.currentTab)
		keyListView := renderKeyListView(m.keys, m.currentKeyIdx, widthKeyListView)
		valueView := renderValueView(m.value, widthValueView)

		return lipgloss.JoinVertical(lipgloss.Center,
			tabRow,
			lipgloss.JoinHorizontal(lipgloss.Top,
				keyListView,
				valueView,
			),
		)
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
}

func renderValueView(value string, width int) string {
	title := "  Value"
	if value == "" {
		value = "No value found for the selected key."
	}

	// Create a styled view for the value
	style := lipgloss.NewStyle().
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width) // Adjust width to fit within the terminal

	return lipgloss.JoinVertical(lipgloss.Top, title, style.Render(value))
}

func renderTabRow(tabs []string, currentTab int) string {
	_tabs := make([]string, len(tabs))
	for i, tab := range tabs {
		if i == currentTab {
			_tabs[i] = activeTabStyle.Render(tab)
		} else {
			_tabs[i] = tabStyle.Render(tab)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, _tabs...)
}

func renderKeyListView(keys []string, cur int, width int) string {
	title := "  Keys"
	style := keyListStyle.PaddingRight(width)

	if len(keys) == 0 {
		return style.Render(list.New([]string{"No keys found"}).String())
	}

	l := list.New(keys).
		ItemStyle(keyListStyle).
		Enumerator(func(items list.Items, i int) string {
			if i == cur {
				return "▶ " // Current item indicator
			}
			return ""
		}).
		ItemStyleFunc(func(items list.Items, i int) lipgloss.Style {
			if i == cur {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("30")).
					Background(lipgloss.Color("44"))
			}
			return lipgloss.NewStyle()
		})

	return lipgloss.JoinVertical(lipgloss.Top, title, style.Render(l.String()))
}

func (m model) centerText(txt string) string {
	padding := max((m.width-len(txt))/2, 0)
	return strings.Repeat(" ", padding) + txt
}

func (m model) UpdateWindowSize(height, width int) model {
	m.width = width
	m.height = height
	// Update key list style with new width
	// keyListStyle = keyListStyle.PaddingRight(width / 3)
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
	m.currentTab = (m.currentTab + 1) % len(m.tabs)
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

func (m model) currentKey() string {
	if m.currentKeyIdx < 0 || m.currentKeyIdx >= len(m.keys) {
		return ""
	}
	return m.keys[m.currentKeyIdx]
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
		tabs:  []string{"GET", "SET", "HGET", "HSET"},
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
