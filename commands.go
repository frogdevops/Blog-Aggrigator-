package main

import (
	"context"
	"fmt"
	"time"

	"github.com/frogdevops/gator/internal/config"
	"github.com/frogdevops/gator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
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
		return handlerRegister(s, cmd)

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
	return command{Name: args[1], Args: args[2:]}, nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: login <username>")
	}
	if err := s.cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("set user: %w", err)
	}
	fmt.Println("user has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: register <name>")
	}

	name := cmd.Args[0]

	dbParam := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	}

	user, err := s.db.CreateUser(context.Background(), dbParam)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	fmt.Printf("user %s created\n", user.Name)
	return nil
}
