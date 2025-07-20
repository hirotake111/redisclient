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
	keys []*values.Key // Stores the keys for the view state
}

func (s ViewState) isState() {}
func (s ViewState) String() string {
	return "view"
}
