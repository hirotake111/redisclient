package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hirotake111/redisclient/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	defaultNumItems      = 1000
	defaultJsonKeyLength = 10
	bigPayloadSize       = 100 // 10KB
)

// This script generates a number of fake data and store it to Redis.
func main() {
	var numItems int
	var bigPayload bool
	flag.IntVar(&numItems, "n", defaultNumItems, "Number of fake items to generate")
	flag.BoolVar(&bigPayload, "b", false, "Generate big payloads (10KB each)")
	flag.Parse()
	fmt.Printf("Generating %d fake items...\n", numItems)

	size := defaultJsonKeyLength
	fmt.Printf("Big	 payload: %v\n", bigPayload)
	if bigPayload {
		size = bigPayloadSize
	}

	ctx := context.Background()

	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		fmt.Printf("Failed to get config from environment: %v\n", err)
		os.Exit(1)
	}

	r := redis.NewClient(cfg.Option)
	if _, err := r.Ping(ctx).Result(); err != nil {
		fmt.Printf("Failed to connect to Redis at %s - %v\n", cfg.Option.Addr, err)
		os.Exit(1)
	}

	fmt.Println("Adding fake data to Redis...")
	for i := 0; i < numItems; i++ {
		key := fmt.Sprintf("key:%d", i)
		value, err := generateJsonPaylod(i, size)
		if err != nil {
			fmt.Printf("Failed to generate JSON payload - %v\n", err)
			os.Exit(1)
		}
		if err := r.Set(ctx, key, value, 0).Err(); err != nil {
			fmt.Printf("Failed to set key %s - %v\n", key, err)
			os.Exit(1)
		}
	}
	fmt.Println("Fake data generated successfully")

}

func generateJsonPaylod(i, size int) (string, error) {
	var sb strings.Builder
	var err error
	for j := range size {
		_, err = sb.WriteString(fmt.Sprintf("\"key%d\":\"value%d%d\"", j, i, j))
		if j < size-1 {
			_, err = sb.WriteString(",")
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to generate JSON payload: %v", err)
	}
	return "{" + sb.String() + "}", nil
}
