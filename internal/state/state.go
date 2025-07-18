package state

type State interface {
	isState()
}

type initialState string

func (s initialState) isState() {}

const InitialState initialState = "initial"

type formState string

func (s formState) isState() {}

const FormState formState = "form"
