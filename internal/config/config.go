package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	appName        = "redisclient"
	configFileName = "config.json"
)

type Config struct {
	Connections []Connection `json:"connections"`
}

type Connection struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DB       int    `json:"db"`
	Password string `json:"password"`
}

func (c *Connection) URL() string {
	return fmt.Sprintf("redis://%s:%d/%d", c.Host, c.Port, c.DB)
}

func GetConfig() *Config {
	// Load base directory
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Failed to get user config directory: %v", err)
	}

	// Create the directory if it doesn't exist
	apath := filepath.Join(dir, appName)
	if _, err := os.Stat(apath); err != nil {
		if err = os.Mkdir(apath, 0755); err != nil {
			log.Fatal(err)
		}
	}

	// Create config file if it doesn't exist
	cpath := filepath.Join(apath, configFileName)
	cfg := new(Config)
	if _, err := os.Stat(cpath); err != nil {
		f, err := os.OpenFile(cpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			log.Fatal(err)
		}
		b, err := json.Marshal(cfg)
		if err != nil {
			log.Fatalf("error marshaling config: %v", err)
		}
		if _, err := f.Write(b); err != nil {
			log.Fatalf("error writing config: %v", err)
		}
	}

	return cfg
}
