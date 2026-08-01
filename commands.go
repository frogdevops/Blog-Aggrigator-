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

func parseCommands(args []string) (command, error) {
	if len(args) < 2 {
		return command{}, fmt.Errorf("usage: gator <command> [args...]")
	}
	return command{Name: args[1], Args: args[:2]}, nil
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
