package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type model struct {
	width  int
	height int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	// Center the message in the window
	msg := "Hello, World!"
	info := "Press any key to exit."

	// Calculate vertical padding
	padTop := (m.height - 2) / 2
	padLeft := (m.width - len(msg)) / 2
	if padTop < 0 {
		padTop = 0
	}
	if padLeft < 0 {
		padLeft = 0
	}

	view := ""
	for i := 0; i < padTop; i++ {
		view += "\n"
	}
	view += fmt.Sprintf("%s%s\n", spaces(padLeft), msg)
	view += fmt.Sprintf("%s%s", spaces(padLeft), info)
	return view
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%*s", n, "")
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}
