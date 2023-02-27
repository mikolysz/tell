package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

func main() {
	// Set up the environment
	env, err := initEnvironment(os.Args[1:])
	must("", err)

	// We don't validate the environment yet, as we expect new, confused tell users to try running it without arguments, which would ordinarily fail.
	// We don't want to show them a "usage" message, but instead a "no token" message.
	// So, we load the config and see if the token exists first

	configPath, err := defaultConfigPath()
	must("Could not get default config path:", err)

	cfg, err := loadConfig(configPath)

	// Crash the program if loading the config has failed, but not if it's just not there
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		must("Could not load config:", err)
	}

	if errors.Is(err, os.ErrNotExist) && env.token == "" {
		var errorMessage = "no bot token found.\n\nTo obtain one, create a bot by sending /newbot to @BotFather (https://t.me/botfather).\n\nSet your token with 'tell -t <token>"
		fmt.Fprintln(os.Stderr, errorMessage)
		os.Exit(1)
	}

	if cfg != nil && cfg.ChatID == 0 && !env.authorizeNewUser {
		fmt.Fprintln(os.Stderr, "No authorized user found. Please authorize a user with 'tell -a'")
		os.Exit(1)
	}

	must("Error:", env.validate())

	// Set the token or authorize the user if requested
	if env.token != "" {
		// It's possible that the config doesn't exist yet, so we need to create it
		if cfg == nil {
			cfg = &config{}
		}

		cfg.BotToken = env.token
		must("Could not save config:", cfg.save(configPath))
		fmt.Fprintf(os.Stderr, "Token has been set!\n\nNow, authorize your Telegram account to receive notifications with 'tell -a'")
		os.Exit(0)
	}

	bot, err := gotgbot.NewBot(cfg.BotToken, nil)
	must("Could not create bot instance:", err)

	if env.authorizeNewUser {
		chatId, err := authorize(bot)
		must("Could not authorize user:", err)
		cfg.ChatID = chatId
		must("Could not save config:", cfg.save(configPath))
		os.Exit(0)
	}

	// If we got here, we're ready to send a message
	_, err = bot.SendMessage(cfg.ChatID, env.message, nil)
	must("Could not send message:", err)
}

func authorize(bot *gotgbot.Bot) (chatId int64, err error) {
	// Generate a random 6-digit code
	code := rand.Intn(999999-100000) + 100000
	fmt.Fprintf(os.Stderr, "Please send the following code to @%s: %d", bot.Username, code)

	chatIdChan := make(chan int64)

	// Prepare to receive messages:
	updater := ext.NewUpdater(nil)
	handler := handlers.NewMessage(message.Contains(fmt.Sprintf("%d", code)), func(b *gotgbot.Bot, ctx *ext.Context) error {
		chatId = ctx.EffectiveChat.Id
		chatIdChan <- chatId
		return nil
	})

	updater.Dispatcher.AddHandler(handler)

	go updater.StartPolling(bot, nil)
	chatId = <-chatIdChan
	updater.Stop()
	return chatId, nil
}

func must(message string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, message, err)
		os.Exit(1)
	}
}
