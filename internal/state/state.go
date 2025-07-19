package state

import "github.com/hirotake111/redisclient/internal/values"

type State interface {
	isState()
}

type initialState string

func (s initialState) isState() {}

const InitialState initialState = "initial"

// form state
type formState string

func (s formState) isState() {}

const FormState formState = "form"

// view state
type viewState struct {
	keys []*values.Key // Stores the keys for the view state
}

var ViewState viewState = viewState{
	keys: make([]*values.Key, 0),
}

func (s viewState) isState() {}
