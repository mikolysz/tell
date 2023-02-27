package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	flag "github.com/spf13/pflag"
)

func main() {
	flag.Parse()

	// Load config
	configPath, err := defaultConfigPath()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create a new bot instance
	bot, err := gotgbot.NewBot(cfg.BotToken, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get the message from the command line
	// This way, we can use tell like echo, without having to quote the message
	message := strings.Join(flag.Args(), " ")

	_, err = bot.SendMessage(cfg.ChatID, message, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
