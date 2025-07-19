package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/config"
	"github.com/hirotake111/redisclient/internal/logger"
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
	message  string        // temporary message for display
}

func CreateInitialModel(r *redis.Client, message string) model {
	return model{
		width:    0, // Default width
		height:   0, // Default height
		redisKey: "",
		state:    state.InitialState, // Start in initial state
		redis:    r,
		message:  message,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.UpdateWindowSize(msg.Height, msg.Width), nil

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
	}

	return fmt.Sprintf("Unknown state: %s", m.state)
}

func (m model) centerText(txt string) string {
	padding := max((m.width-len(txt))/2, 0)
	return strings.Repeat(" ", padding) + txt
}

func (m model) toFormState() model {
	m.state = state.FormState // Transition to form state
	return m
}
func (m model) toViewState() model {
	m.state = state.ViewState // Transition to view state
	return m
}

func (m model) toInitialState() model {
	m.state = state.InitialState // Transition back to initial state
	return m
}

func (m model) UpdateWindowSize(height, width int) model {
	m.width = width
	m.height = height
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

func main() {
	// Initialize logger to write to temp file
	if err := logger.InitLogger(); err != nil {
		// If logger fails, print to stderr and exit
		log.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Print("Logger initialized successfully")

	ctx := context.Background()

	_ = config.GetConfig()

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
		os.Exit(1)
	}

	m := CreateInitialModel(r, "Welcome to the Redis Client!")
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	log.Print("Program exited successfully")
}
