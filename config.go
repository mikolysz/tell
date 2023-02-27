package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type config struct {
	BotToken string `json:"bot_token"`
	ChatID   int64  `json:"chat_id"`
}

func defaultConfigPath() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return dirname + "/.tell.json", nil
}

func loadConfig(path string) (*config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}
