package main

import (
	"fmt"

	"github.com/frogdevops/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	registry map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	switch cmd.Name {
	case "login":
		return handlerLogin(s, cmd)
	case "register":

	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registry[name] = f
}

func parseCommands(args []string) (command, error) {
	if len(args) < 2 {
		return command{}, fmt.Errorf("usage: gator <command> [args...]")
	}
	fmt.Println("user has been set")
	return command{Name: args[1], Args: args[2:]}, nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}
	if err := s.cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("set user: %w", err)
	}
	return nil
}
