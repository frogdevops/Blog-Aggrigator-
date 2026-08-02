package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/frogdevops/gator/internal/config"
	"github.com/frogdevops/gator/internal/database"
)

import _ "github.com/lib/pq"

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	dbQueries := database.New(db)

	s := &state{db: dbQueries, cfg: &cfg}

	cmds := commands{registry: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddfeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	cmd, err := parseCommands(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	if err := cmds.run(s, cmd); err != nil {
		log.Fatal(err)
	}
	_ = s
}
