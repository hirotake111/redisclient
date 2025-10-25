package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hirotake111/redisclient/internal/config"
	"github.com/redis/go-redis/v9"
)

const (
	defaultNumItems = 1000
)

// This script generates a number of fake data and store it to Redis.
func main() {
	var numItems int
	flag.IntVar(&numItems, "n", defaultNumItems, "Number of fake items to generate")
	flag.Parse()
	fmt.Printf("Generating %d fake items...\n", numItems)

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
		value := fmt.Sprintf("value:%d", i)
		if err := r.Set(ctx, key, value, 0).Err(); err != nil {
			fmt.Printf("Failed to set key %s - %v\n", key, err)
			os.Exit(1)
		}
	}
	fmt.Println("Fake data generated successfully")

}
