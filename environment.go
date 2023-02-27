package main

import (
	"fmt"
	"strings"

	flag "github.com/spf13/pflag"
)

// environment contains the flags and arguments passed to the program
type environment struct {
	fs *flag.FlagSet

	token            string
	authorizeNewUser bool

	// any remaining arguments are concatenated into a single string,
	// so that we can use tell like echo, without having to quote the message
	message string
}

func initEnvironment(args []string) (*environment, error) {
	// set up and parse the flags

	// We continue on error here so tests can run
	// main catches the error, and we exit there
	fs := flag.NewFlagSet("tell", flag.ContinueOnError)
	env := &environment{fs: fs}

	fs.StringVarP(&env.token, "token", "t", "", "Save the provided Telegram bot token in the config file")
	fs.BoolVarP(&env.authorizeNewUser, "authorize-user", "a", false, "Authorize a new user to use the bot")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Get the message from the command line
	// This way, we can use tell like echo, without having to quote the message
	env.message = strings.Join(fs.Args(), " ")

	return env, nil
}

func (env *environment) validate() error {
	// If any of the config-modifying flags were provided, message should be empty
	if env.token != "" || env.authorizeNewUser {
		if env.message != "" {
			return fmt.Errorf("both a message and a config-modifying flag were provided")
		}
		return nil
	}

	return nil
}
