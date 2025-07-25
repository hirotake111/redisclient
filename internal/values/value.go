package values

type Value struct {
	data string // The value data
	ttl  int64  // Time to live for the value
}

func NewValue(data string, ttl int64) Value {
	return Value{
		data: data,
		ttl:  ttl,
	}
}

func (v Value) Data() string {
	return v.data
}
func (v Value) TTL() int64 {
	return v.ttl
}
