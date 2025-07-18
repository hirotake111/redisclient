package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type model struct {
	width    int
	height   int
	redisKey string // Stores the Redis key input
	state    string // "initial" or "form"
}

func (m model) Init() tea.Cmd {
	m.state = "initial"
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if m.state == "initial" {
			if key == "enter" {
				m.state = "form"
				return m, nil
			}
			// Ignore other keys in initial state
			return m, nil
		}
		if m.state == "form" {
			if key == "enter" {
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	if m.state == "initial" {
		return "hello world\n(Press Enter to continue)"
	}
	// Form view
	label := "Enter Redis key:"
	input := m.redisKey
	info := "Type your Redis key and press Enter. Backspace deletes."

	view := ""
	view += fmt.Sprintf("%s %s\n", label, input)
	view += info
	return view
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}

func main() {
	p := tea.NewProgram(model{state: "initial"}, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}
