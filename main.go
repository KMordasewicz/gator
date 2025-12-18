package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/KMordasewicz/gator/internal/config"
	"github.com/KMordasewicz/gator/internal/database"
	_ "github.com/lib/pq"
)

func registerHandles(c *commands) {
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handleReset)
	c.register("users", handleUsers)
	c.register("agg", hanldeAgg)
	c.register("addfeed", middlewareLoggedIn(handleAddFeed))
	c.register("feeds", handleFeeds)
	c.register("follow", middlewareLoggedIn(handleFollow))
	c.register("following", middlewareLoggedIn(handleFollowing))
	c.register("unfollow", middlewareLoggedIn(handleUnfollow))
	c.register("browse", middlewareLoggedIn(handleBrowse))
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Couldn't read config: %e\n", err)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to postgres: %s, due to error: %e\n", cfg.DbUrl, err)
	}
	dbQueries := database.New(db)
	s := state{db: dbQueries, cfg: &cfg}
	c := commands{callable: make(map[string]func(*state, command) error)}
	registerHandles(&c)
	if len(os.Args) < 2 {
		log.Fatal("Insufficient number of arguments provided.\n")
	}
	commandName := os.Args[1]
	args := os.Args[2:]
	cmd := command{name: commandName, args: args}
	err = c.run(&s, cmd)
	if err != nil {
		log.Fatalf("Failed to run a %s command: %v\n", commandName, err)
	}
}
