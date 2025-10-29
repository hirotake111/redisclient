package command

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/redis/go-redis/v9"
)

func DisplayEmptyValue() tea.Msg {
	return ValueUpdatedMsg{}
}

func GetKeys(ctx context.Context, redis *redis.Client, pattern string) tea.Cmd {
	return func() tea.Msg {
		if pattern == "" {
			pattern = "*"
		}

		log.Printf("Fetching keys from Redis with pattern \"%s\", db: %d", pattern, redis.Options().DB)
		keys, err := redis.Keys(ctx, pattern).Result()
		if err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Fetched %d keys from Redis(DB: %d)", len(keys), redis.Options().DB)

		return KeysUpdatedMsg{Keys: keys}
	}
}

func GetValue(ctx context.Context, redis *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		t, err := redis.Type(ctx, key).Result()
		if err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Fetching value for key \"%s\" of type %s", key, t)
		var newValue string
		switch t {
		case "string":
			value, err := redis.Get(ctx, key).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			log.Printf("Fetched value for key \"%s\"", key)
			newValue = escapeCharacter(value)

		case "hash":
			hm, err := redis.HGetAll(ctx, key).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			bytes, err := json.Marshal(hm)
			if err != nil {
				return ErrMsg{Err: err}
			}
			newValue = (string(bytes))

		case "list":
			list, err := redis.LRange(ctx, key, 0, -1).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			bytes, err := json.Marshal(list)
			if err != nil {
				return ErrMsg{Err: err}
			}
			newValue = (string(bytes))

		case "set":
			members, err := redis.SMembers(ctx, key).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			bytes, err := json.Marshal(members)
			if err != nil {
				return ErrMsg{Err: err}
			}
			newValue = (string(bytes))

		case "zset":
			zset, err := redis.ZRangeWithScores(ctx, key, 0, -1).Result()
			if err != nil {
				return ErrMsg{Err: err}
			}
			// Convert ZSet to a map for easier display
			zsetMap := make(map[string]float64)
			for _, z := range zset {
				zsetMap[z.Member.(string)] = z.Score
			}
			bytes, err := json.Marshal(zsetMap)
			if err != nil {
				return ErrMsg{Err: err}
			}
			newValue = (string(bytes))

		case "none": // Key does not exist
			newValue = ""
			return ErrMsg{Err: fmt.Errorf("key %s does not exist in the database", key)}

		default:
			return ErrMsg{Err: fmt.Errorf("unsupported type %s for key %s", t, key)}
		}

		log.Printf("Fetching TTL for key %s of type %s", key, t)
		ttl, err := redis.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("Error fetching TTL for key %s: %v", key, err)
		}
		return ValueUpdatedMsg{
			NewValue: newValue,
			TTL:      int64(ttl.Seconds()), // Convert TTL to seconds
		}
	}
}

func escapeCharacter(value string) string {
	runes := make([]rune, 0, len(value))
	for _, r := range value {
		if r >= 32 {
			runes = append(runes, r)
		}
	}
	return string(runes)
}

func UpdateValue(ctx context.Context, client *redis.Client, key string, newValue string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Updating key %s with new value %s", key, newValue)
		if err := client.Set(ctx, key, newValue, 0).Err(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Updated key %s successfully", key)
		return ValueUpdatedMsg{
			NewValue: newValue,
		}
	}
}

func DeleteKey(ctx context.Context, client *redis.Client, key string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Deleting key \"%s\" from Redis", key)
		if err := client.Del(ctx, key).Err(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Deleted key \"%s\" successfully", key)
		return KeyDeletedMsg{Key: key, info: "Key deleted successfully"}
	}
}

func SwitchTab(ctx context.Context, client *redis.Client, tab int) tea.Cmd {
	log.Printf("Switching to tab %d", tab)
	client.Options().DB = tab
	nc := redis.NewClient(client.Options())

	return func() tea.Msg {
		if _, err := nc.Ping(ctx).Result(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Switched to tab %d", tab)
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
		// Currently only supports macOS (pbcopy)
		command := exec.Command("pbcopy")
		command.Stdin = strings.NewReader(value)
		if err := command.Run(); err != nil {
			return ErrMsg{Err: fmt.Errorf("failed to copy value to clipboard: %w", err)}
		}
		log.Print("Value copied to clipboard successfully")
		return CopySuccessMsg{}
	}
}

// TickAndClear creates a command that ticks every duration and returns a TimedOutMsg.
func TickAndClear(duration time.Duration, kind string) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return TimedOutMsg{Kind: kind}
	})

}

func UpdateSelectedItemCmd(newKey string) tea.Msg {
	return HighlightedKeyUpdatedMsg{}
}

func BulkDelete(ctx context.Context, client *redis.Client, keys []string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Bulk deleting %d keys from Redis", len(keys))
		if err := client.Del(ctx, keys...).Err(); err != nil {
			return ErrMsg{Err: err}
		}
		log.Printf("Bulk deleted %d keys successfully", len(keys))

		// Get new values for refreshing the key list
		return GetKeys(ctx, client, "")()
	}
}
