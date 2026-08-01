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
	handler, ok := c.registry[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
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
	name := cmd.Args[0]
	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user %s not found: %w", name, err)
	}
	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("set user: %w", err)
	}

	fmt.Printf("logged in as %s\n", user.Name)
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

func handlerReset(s *state, cmd command) error {
	if err := s.db.Reset(context.Background()); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	fmt.Println("all users has been deleted.")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("get users: %w", err)
	}

	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}
