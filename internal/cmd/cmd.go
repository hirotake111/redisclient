package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
)

const (
	keysPreQuery = 40 // Number of keys to prefetch when scanning Redis
)

type ErrMsg struct{ Err error }

type KeysUpdatedMsg struct {
	Keys        []string
	RedisCursor uint64
}

type ValueMsg string

type NewRedisClientMsg struct {
	Redis *redis.Client
}

type KeyDeletedMsg struct {
	Key string
}

type CopySuccessMsg struct{}

func DisplayEmptyValue() tea.Msg {
	return ValueMsg("")
}

func GetKeys(ctx context.Context, redis *redis.Client, cursor uint64) tea.Cmd {
	return func() tea.Msg {
		log.Print("Fetching keys from Redis...")
		// keys, err := redis.Keys(ctx, "*").Result()
		keys, cursor, err := redis.Scan(ctx, cursor, "", keysPreQuery).Result()
		if err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Fetched %d keys from Redis. Cursor: %d", len(keys), cursor)
		return KeysUpdatedMsg{Keys: keys, RedisCursor: cursor}
	}
}

func GetValue(ctx context.Context, redis *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		t, err := redis.Type(ctx, key).Result()
		if err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Fetching value for key %s of type %s", key, t)
		switch t {
		case "string":
			value, err := redis.Get(ctx, key).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			log.Printf("Fetched value for key %s", key)
			return ValueMsg(value)

		case "hash":
			hm, err := redis.HGetAll(ctx, key).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			bytes, err := json.Marshal(hm)
			if err != nil {
				return ErrMsg{Err: err}
			}
			return ValueMsg(string(bytes))
		}

		return ErrMsg{Err: fmt.Errorf("unsupported type %s for key %s", t, key)}
	}
}

func UpdateValue(ctx context.Context, client *redis.Client, key string, newValue string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Updating key %s with new value %s", key, newValue)
		if err := client.Set(ctx, key, newValue, 0).Err(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Updated key %s successfully", key)
		return ValueMsg(newValue) // Return the new value as a message
	}
}

func DeleteKey(ctx context.Context, client *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Deleting key %s from Redis", key)
		if err := client.Del(ctx, key).Err(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Deleted key %s successfully", key)
		return KeyDeletedMsg{Key: key}
	}
}

func UpdateDatabase(ctx context.Context, client *redis.Client, db int) tea.Cmd {
	log.Printf("Updating Redis database to %d", db)
	client.Options().DB = db
	nc := redis.NewClient(client.Options())
	return func() tea.Msg {
		if _, err := nc.Ping(ctx).Result(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Switched to Redis database %d", db)
		return NewRedisClientMsg{Redis: nc} // No message needed for successful DB switch
	}
}

func CopyValueToClipboard(ctx context.Context, value string) tea.Cmd {
	return func() tea.Msg {
		var truncated = value
		if len(value) > 10 {
			truncated = value[:10] + "..." // Truncate long values for logging
		}
		log.Printf("Copying value to clipboard: %s", truncated)
		// TODO: Implement platform-specific clipboard handling
		command := exec.Command("pbcopy")
		command.Stdin = strings.NewReader(value)
		if err := command.Run(); err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to copy value to clipboard: %w", err)}
		}
		log.Print("Value copied to clipboard successfully")
		return CopySuccessMsg{}
	}
}
