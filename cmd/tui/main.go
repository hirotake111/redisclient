package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hirotake111/redisclient/internal/config"
	"github.com/hirotake111/redisclient/internal/logger"
	"github.com/hirotake111/redisclient/internal/model"
	"github.com/redis/go-redis/v9"
)

var (
	Version = "development"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version number")
	flag.Parse()

	if showVersion != nil && *showVersion {
		fmt.Printf("VERSION: %s\n", Version)
		os.Exit(0)
	}

	// Initialize logger to write to temp file
	if err := logger.Initialize(); err != nil {
		// If logger fails, print to stderr and exit
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	log.Print("Logger initialized successfully")

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

	m := model.NewModel(ctx, r)

	log.Println("Starting app now...")
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithoutBracketedPaste())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Program exited successfully")
}
