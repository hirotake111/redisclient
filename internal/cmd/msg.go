package cmd

import (
	"github.com/redis/go-redis/v9"
)

type ErrMsg struct{ Err error }

type TimedOutMsg struct {
	Kind string // Type of timeout, e.g., "network", "error", "redis", etc.
}

type KeysUpdatedMsg struct {
	Keys        []string
	RedisCursor uint64
}

type ValueUpdatedMsg struct {
	NewValue string // The new value for the key
	TTL      int64  // Time to live for the key, if applicable
}

type NewRedisClientMsg struct {
	Redis *redis.Client
}

type KeyDeletedMsg struct {
	Key string
}

type CopySuccessMsg struct{}
