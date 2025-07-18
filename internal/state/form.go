package state

type Form string

func (f Form) String() string {
	return string(f)
}

func (f Form) RemoveRight() Form {
	return f[:len(f)-1]
}

func (f Form) AppendRight(s string) Form {
	return f + Form(s)
}
