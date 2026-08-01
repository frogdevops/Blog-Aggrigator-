package main

import (
	"log"
	"os"

	"github.com/frogdevops/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	s := &state{cfg: &cfg}

	cmds := commands{registry: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)

	cmd, err := parseCommands(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	if err := cmds.run(s, cmd); err != nil {
		log.Fatal(err)
	}
	_ = s
}
