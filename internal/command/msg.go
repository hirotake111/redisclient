package command

import (
	"github.com/redis/go-redis/v9"
)

type MsgWithKind interface {
	Kind() string
}

type TimedOutMsg struct {
	Kind string // Type of timeout, e.g., "network", "error", "redis", etc.
}

type KeysUpdatedMsg struct {
	Keys []string
}

func (KeysUpdatedMsg) Kind() string {
	return "keys_updated"
}

type ValueUpdatedMsg struct {
	NewValue string // The new value for the key
	TTL      int64  // Time to live for the key, if applicable
}

type NewRedisClientMsg struct {
	Redis *redis.Client
}

type KeyDeletedMsg struct {
	Key  string
	info string
}

type CopySuccessMsg struct{}

type HighlightedKeyUpdatedMsg struct {
	Key string
}
