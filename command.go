package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/KMordasewicz/blog-aggregator/internal/config"
	"github.com/KMordasewicz/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	callable map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.callable[cmd.name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.callable[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing username argument")
	}
	name := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, name)
	if err != nil {
		log.Fatalf("User %s doesn't exists\n", name)
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("Username has been successfully set to %s\n", name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing username argument")
	}
	name := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, name)
	if err == nil {
		log.Fatal("User already exists")
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("User %s was created: %v\n", name, user)
	return nil
}
