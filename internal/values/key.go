package values

type Key string

func (k *Key) String() string {
	return string(*k)
}
