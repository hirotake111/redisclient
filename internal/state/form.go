package state

type Form string

func (f Form) String() string {
	return string(f)
}

func (f Form) RemoveRight() Form {
	if len(f) > 0 {
		return f[:len(f)-1]
	}
	return f
}

func (f Form) AppendRight(s string) Form {
	return f + Form(s)
}
