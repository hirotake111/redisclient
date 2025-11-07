package config

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Option *redis.Options
}

func GetConfigFromEnv() (*Config, error) {
	addr := os.Getenv("REDIS_URL")
	var opt *redis.Options
	var err error
	if addr == "" {
		opt, err = redis.ParseURL("redis://localhost:6379")
	} else {
		opt, err = redis.ParseURL(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse REDIS_URL: %w", err)
	}

	return &Config{Option: opt}, nil
}
