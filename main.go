package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/KMordasewicz/blog-aggregator/internal/config"
	"github.com/KMordasewicz/blog-aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Couldn't read config: %e\n", err)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to postgres: %s, due to error: %e", cfg.DbUrl, err)
	}
	dbQueries := database.New(db)
	s := state{db: dbQueries, cfg: &cfg}
	c := commands{callable: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
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
