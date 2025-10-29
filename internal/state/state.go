package state

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type data string

const (
	ViewportActivated   data = "viewport_activated"
	ViewportDeactivated data = "viewport_deactivated"
)

type AppStateTransitionedMsg struct {
	kind string
	data data
}

func ActivateViewportCmd() tea.Msg {
	log.Println("Sending AppStateTransitionedMsg")
	return AppStateTransitionedMsg{
		kind: string(ViewportActivated),
		data: ViewportActivated,
	}
}
func DeactivateViewportCmd() tea.Msg {
	return AppStateTransitionedMsg{
		kind: string(ViewportDeactivated),
		data: ViewportDeactivated,
	}
}

type AppState struct {
	listActive     bool
	viewportActive bool
}

func NewAppState() AppState {
	return AppState{
		listActive:     true,
		viewportActive: false,
	}
}

func (s AppState) Update(msg tea.Msg) (AppState, tea.Cmd) {
	m, ok := msg.(AppStateTransitionedMsg)
	if !ok {
		return s, nil
	}

	switch m.data {
	case ViewportActivated:
		s.listActive = false
		s.viewportActive = true
	case ViewportDeactivated:
		s.listActive = true
		s.viewportActive = false
	}

	return s, nil
}

func (s AppState) ListActive() bool {
	return s.listActive
}

func (s AppState) ViewportActive() bool {
	return s.viewportActive
}
