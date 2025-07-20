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
	"github.com/hirotake111/redisclient/internal/state"
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

type model struct {
	ctx        context.Context // Context for app
	tabs       []string        // List of tabs
	currentTab int             // Current tab index
	width      int             // Width of the terminal window
	height     int             // Height of the terminal window
	redisKey   state.Form      // Stores the Redis key input
	state      state.State     // "initial" or "form"
	redis      *redis.Client   // Redis client instance
	message    string          // temporary message for display
	value      string          // Stores the value for the current key
}

func (m model) Init() tea.Cmd {
	log.Print("Initializing model...")
	return cmd.GetKeys(m.ctx, m.redis)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch st := m.state.(type) {
	case state.InitialState:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			return m.UpdateWindowSize(msg.Height, msg.Width), nil
		case tea.KeyMsg:
			key := msg.String()
			if key == "enter" {
				return m.toFormState(), nil
			}
			if key == "esc" || key == "ctrl+c" || key == "ctrl+q" {
				return m, tea.Quit
			}
			// Ignore other keys in initial state
			return m, nil
		}

	case state.FormState:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			return m.UpdateWindowSize(msg.Height, msg.Width), nil
		case tea.KeyMsg:
			key := msg.String()
			if key == "enter" {
				// TODO: Use the redisKey for some operation
				return m, tea.Quit
			}
			if key == "backspace" || key == "ctrl+h" {
				if len(m.redisKey) > 0 {
					m.redisKey = m.redisKey[:len(m.redisKey)-1]
				}
				return m, nil
			}
			if len(key) == 1 && msg.Type == tea.KeyRunes {
				return m.AppendRedisKey(key), nil
			}
		}

	case state.ViewState:
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
				log.Printf("Current index before moving down: %d", st.Current)
				st = st.MoveCursorDown()
				m.state = st
				log.Printf("Current index after moving down: %d", st.Current)
				log.Printf("Keys: %v", st.Keys)
				return m, cmd.GetValue(m.ctx, m.redis, st.Keys[st.Current])
			}
			if key == "k" {
				log.Print("Moving cursor up")
				st = st.MoveCursorUp()
				m.state = st
				return m, cmd.GetValue(m.ctx, m.redis, st.Keys[st.Current])
			}
			if key == "tab" {
				// Switch to next tabIndex
				m = m.NextTab()
				return m, cmd.GetValue(m.ctx, m.redis, st.Keys[st.Current])
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
	switch st := m.state.(type) {
	case state.InitialState:
		paddingHeight := max(0, m.height/2)
		verticalPadding := strings.Repeat("\n", paddingHeight)
		info := fmt.Sprintf("height: %d, width: %d, message: %s\n", m.height, m.width, m.message)
		line1 := "Welcome to the Redis Client!"
		line2 := "Press Enter to start, or Esc to quit."

		return verticalPadding +
			m.centerText(line1) + "\n" +
			m.centerText(line2) + "\n" +
			m.centerText(info) + "\n"

	case state.FormState:
		// Form view
		label := "Enter Redis key:"
		input := m.redisKey
		info := "Type your Redis key and press Enter. Backspace deletes."

		view := ""
		view += fmt.Sprintf("%s %s\n", label, input)
		view += info
		return view

	case state.ViewState:
		tabRow := renderTabRow(m.tabs, m.currentTab)
		klv := renderKeyListView(st, m.width)
		valueView := renderValueView(m.value, m.width)
		return lipgloss.JoinVertical(lipgloss.Center,
			tabRow,
			lipgloss.JoinHorizontal(lipgloss.Top,
				klv,
				valueView,
			),
		)
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
}

func renderValueView(value string, width int) string {
	if value == "" {
		value = "No value found for the selected key."
	}

	// Create a styled view for the value
	style := lipgloss.NewStyle().
		Padding(1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(width/3 - 2) // Adjust width to fit within the terminal

	return style.Render(value)
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

func renderKeyListView(v state.ViewState, width int) string {
	style := keyListStyle.PaddingRight(width / 3)

	if len(v.Keys) == 0 {
		return style.Render(list.New([]string{"No keys found"}).String())
	}

	ks := make([]string, 0, len(v.Keys))
	for _, key := range v.Keys {
		ks = append(ks, key.String())
	}

	l := list.New(ks).
		ItemStyle(keyListStyle).
		Enumerator(func(items list.Items, i int) string {
			if i == v.Current {
				return "▶ " // Current item indicator
			}
			return ""
		}).
		ItemStyleFunc(func(items list.Items, i int) lipgloss.Style {
			if i == v.Current {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("30")).
					Background(lipgloss.Color("44"))
			}
			return lipgloss.NewStyle()
		})

	return style.Render(l.String())
}

func (m model) centerText(txt string) string {
	padding := max((m.width-len(txt))/2, 0)
	return strings.Repeat(" ", padding) + txt
}

func (m model) toFormState() model {
	m.state = state.FormState("") // Transition to form state
	return m
}
func (m model) toViewState() model {
	m.state = state.ViewState{} // Transition to view state
	return m
}

func (m model) UpdateWindowSize(height, width int) model {
	m.width = width
	m.height = height
	// Update key list style with new width
	// keyListStyle = keyListStyle.PaddingRight(width / 3)
	return m
}

func (m model) AppendRedisKey(s string) model {
	m.redisKey = m.redisKey.AppendRight(s)
	return m
}

func (m model) RemoveRightRedisKey() model {
	m.redisKey = m.redisKey.RemoveRight()
	return m
}

func (m model) UpdateKeyList(msg cmd.KeysUpdatedMsg) model {
	m.state = state.ViewState{Keys: msg}
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
		state: state.ViewState{},
		redis: r,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	log.Print("Program exited successfully")
}
