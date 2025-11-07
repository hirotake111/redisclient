package command

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type MsgWithKind interface {
	String() string
}

type KeysUpdatedMsg struct {
	Keys []string
}

func (KeysUpdatedMsg) String() string {
	return "keys_updated"
}

type ValueUpdatedMsg struct {
	NewValue string // The new value for the key
	TTL      int64  // Time to live for the key, if applicable
}

func (v ValueUpdatedMsg) String() string {
	return fmt.Sprintf("value_updated (TTL: %d)", v.TTL)
}

type NewRedisClientMsg struct {
	Redis *redis.Client
}

func (n NewRedisClientMsg) String() string {
	return "new_redis_client"
}

type KeyDeletedMsg struct {
	Key  string
	info string
}

func (k KeyDeletedMsg) String() string {
	return fmt.Sprintf("key_deleted - key: %s, info: %s", k.Key, k.info)
}

type CopySuccessMsg struct{}

type HighlightedKeyUpdatedMsg struct {
	Key string
}

func (h HighlightedKeyUpdatedMsg) String() string {
	return fmt.Sprintf("highlighted_key_updated - key: %s", h.Key)
}

type TickMsg struct {
	Time time.Time
}

func (t TickMsg) String() string {
	return fmt.Sprintf("tick - time: %s", t.Time.String())
}
