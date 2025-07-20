package state

import "github.com/hirotake111/redisclient/internal/values"

type State interface {
	isState()
}

type InitialState string

func (s InitialState) isState() {}

// form state
type FormState string

func (s FormState) isState() {}

// view state
type ViewState struct {
	Keys    []*values.Key // Stores the keys for the view state
	Current int           // Current index in the keys slice
}

func (s ViewState) isState() {}
func (s ViewState) String() string {
	return "view"
}

func (s ViewState) MoveCursorDown() ViewState {
	if s.Current < len(s.Keys)-1 {
		s.Current++
	}
	return s
}

func (s ViewState) MoveCursorUp() ViewState {
	if s.Current > 0 {
		s.Current--
	}
	return s
}
