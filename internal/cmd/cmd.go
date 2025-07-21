package cmd

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
)

type ErrMsg struct{ err error }

type KeysUpdatedMsg []string

type ValueMsg string

type NewRedisClientMsg struct {
	Redis *redis.Client
}

func GetKeys(ctx context.Context, redis *redis.Client) tea.Cmd {
	log.Print("Fetching keys from Redis...")
	return func() tea.Msg {
		log.Print("Executing Redis KEYS command...")
		keys, err := redis.Keys(ctx, "*").Result()
		if err != nil {
			return ErrMsg{err: err}
		}
		log.Printf("Fetched %d keys from Redis", len(keys))
		return KeysUpdatedMsg(keys)
	}
}

func GetValue(ctx context.Context, redis *redis.Client, keys []string) tea.Cmd {
	if len(keys) == 0 {
		// Empty value display
		return func() tea.Msg { return ValueMsg("") }
	}

	return func() tea.Msg {
		log.Printf("Fetching value for key: %s", keys)
		value, err := redis.Get(ctx, keys[0]).Result()
		if err != nil {
			return ErrMsg{err: err}
		}
		log.Printf("Fetched value for key %s: %s", keys, value)
		return ValueMsg(value)
	}
}

func UpdateDatabase(ctx context.Context, client *redis.Client, db int) tea.Cmd {
	log.Printf("Updating Redis database to %d", db)
	client.Options().DB = db
	nc := redis.NewClient(client.Options())
	return func() tea.Msg {
		if _, err := nc.Ping(ctx).Result(); err != nil {
			return ErrMsg{err: err}
		}
		log.Printf("Switched to Redis database %d", db)
		return NewRedisClientMsg{Redis: nc} // No message needed for successful DB switch
	}
}
