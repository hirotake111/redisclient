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

//	var ViewportActivatedMsg AppStateTransitionedMsg = AppStateTransitionedMsg{
//		kind: "viewport_activated",
//		data: ViewportActivated,
//	}
var ViewportDeactivatedMsg AppStateTransitionedMsg = AppStateTransitionedMsg{
	kind: "viewport_activated",
	data: ViewportDeactivated,
}

func ActivateViewportCmd() tea.Msg {
	log.Println("Sending AppStateTransitionedMsg")
	return AppStateTransitionedMsg{
		kind: "viewport_activated",
		data: ViewportActivated,
	}
}
func DeactivateViewportCmd() tea.Msg {
	return AppStateTransitionedMsg{
		kind: "viewport_activated",
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
	log.Printf("AppState received message: %+v", msg)
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
	log.Printf("Current	AppState: %+v", s)
	return s, nil
}

func (s AppState) ListActive() bool {
	return s.listActive
}

func (s AppState) ViewportActive() bool {
	return s.viewportActive
}
