package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/state"
)

type model struct {
	width    int
	height   int
	redisKey string      // Stores the Redis key input
	state    state.State // "initial" or "form"
}

func CreateInitialModel() model {
	return model{
		width:    80, // Default width
		height:   24, // Default height
		redisKey: "",
		state:    state.InitialState, // Start in initial state
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
				m.redisKey += key
				return m, nil
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
	}
}

func (m model) UpdateWindowSizee(height, width int) model {
	return model{
		width:    width,
		height:   height,
		redisKey: m.redisKey,
		state:    m.state, // Keep the current state
	}
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}

func main() {
	m := CreateInitialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}
