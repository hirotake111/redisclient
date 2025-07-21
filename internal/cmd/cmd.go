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

func GetValue(ctx context.Context, redis *redis.Client, key string) tea.Cmd {
	log.Printf("Fetching value for key: %s", key)
	return func() tea.Msg {
		value, err := redis.Get(ctx, key).Result()
		if err != nil {
			return ErrMsg{err: err}
		}
		log.Printf("Fetched value for key %s: %s", key, value)
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
