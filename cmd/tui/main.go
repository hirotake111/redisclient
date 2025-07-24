package main

import (
	"context"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/config"
	"github.com/hirotake111/redisclient/internal/logger"
	"github.com/hirotake111/redisclient/internal/model"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize logger to write to temp file
	if err := logger.Initialize(); err != nil {
		// If logger fails, print to stderr and exit
		log.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Print("Logger initialized successfully")

	ctx := context.Background()

	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		log.Printf("Failed to get config from environment: %v\n", err)
		os.Exit(1)
	}

	r := redis.NewClient(cfg.Option)
	if _, err := r.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s - %v", cfg.Option.Addr, err)
		os.Exit(1)
	}

	m := model.NewModel(ctx, r)

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithoutBracketedPaste())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	log.Print("Program exited successfully")
}
