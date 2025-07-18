package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/state"
	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisURL = "redis://localhost:6379"
)

type model struct {
	width    int
	height   int
	redisKey state.Form    // Stores the Redis key input
	state    state.State   // "initial" or "form"
	redis    *redis.Client // Redis client instance
}

func CreateInitialModel(r *redis.Client) model {
	return model{
		width:    0, // Default width
		height:   0, // Default height
		redisKey: "",
		state:    state.InitialState, // Start in initial state
		redis:    r,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.UpdateWindowSizee(msg.Height, msg.Width), nil

	case tea.KeyMsg:
		key := msg.String()
		switch m.state {
		case state.InitialState:
			if key == "enter" {
				return m.toFormState(), nil
			}
			if key == "esc" || key == "ctrl+c" || key == "ctrl+q" {
				return m, tea.Quit
			}
			// Ignore other keys in initial state
			return m, nil

		case state.FormState:
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
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case state.InitialState:
		paddingHeight := max(0, m.height/2)
		verticalPadding := strings.Repeat("\n", paddingHeight)
		info := fmt.Sprintf("height: %d, width: %d\n", m.height, m.width)
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
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
}

func (m model) centerText(txt string) string {
	padding := max((m.width-len(txt))/2, 0)
	return spaces(padding) + txt
}

func (m model) toFormState() model {
	return model{
		width:    m.width,
		height:   m.height,
		redisKey: "",
		state:    state.FormState, // Transition to form state
	}
}

func (m model) toInitialState() model {
	return model{
		width:    m.width,
		height:   m.height,
		redisKey: "",
		state:    state.InitialState, // Transition back to initial state
		redis:    m.redis,
	}
}

func (m model) UpdateWindowSizee(height, width int) model {
	return model{
		width:    width,
		height:   height,
		redisKey: m.redisKey,
		state:    m.state, // Keep the current state
		redis:    m.redis,
	}
}

func (m model) AppendRedisKey(s string) model {
	return model{
		width:    m.width,
		height:   m.height,
		redisKey: m.redisKey.AppendRight(s),
		state:    m.state,
		redis:    m.redis,
	}
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}

func main() {
	ctx := context.Background()
	addr := os.Getenv("REDIS_URL")
	if addr == "" {
		addr = defaultRedisURL
	}
	password := os.Getenv("REDIS_PASSWORD")
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	if _, err := r.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s - %v", addr, err)
	}
	m := CreateInitialModel(r)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
