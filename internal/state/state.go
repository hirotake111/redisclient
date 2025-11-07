package state

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type data string

const (
	ViewportActivated   data = "viewport_activated"
	ViewportDeactivated data = "viewport_deactivated"
)

type AppStateTransitionedMsg struct {
	data data
}

func (m AppStateTransitionedMsg) Data() string {
	return string(m.data)
}

func (m AppStateTransitionedMsg) String() string {
	return fmt.Sprintf("app_state_transitioned - data: '%s'", m.Data())
}

func ActivateViewportCmd() tea.Msg {
	return AppStateTransitionedMsg{
		data: ViewportActivated,
	}
}
func DeactivateViewportCmd() tea.Msg {
	return AppStateTransitionedMsg{
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
	log.Printf("AppState received AppStateTransitionedMsg: %+v", m)
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
