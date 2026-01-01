package main

import (
	"context"
	"database/sql"
	"fmt"
	"grysha11/BlogAggregator/internal/config"
	"grysha11/BlogAggregator/internal/database"
	"time"

	"github.com/google/uuid"
)

type state struct {
	db 		*database.Queries
	config	*config.Config
}

type command struct {
	name		string
	args		[]string
}

type commands struct {
	knownCommands	map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	var err error

	if fun, ok := c.knownCommands[cmd.name]; ok {
		err = fun(s, cmd)
	} else {
		err = fmt.Errorf("unknown command")
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.knownCommands == nil {
		c.knownCommands = make(map[string]func(*state, command) error)
	}
	c.knownCommands[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: %v <%v>\nUsage: login <argument>", cmd.name, cmd.args)
	}

	checkExist, err := s.db.GetUser(context.Background(), cmd.args[0])
	if checkExist.ID == uuid.Nil {
		return fmt.Errorf("user don't exist")
	}
	if err != nil {
		return err
	}

	if err := s.config.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Printf("Username have been set to %v\n", s.config.CurrentUsername)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: %v <%v>\nUsage: register <argument>", cmd.name, cmd.args)
	}

	checkDup, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil && checkDup.ID != uuid.Nil {
		return fmt.Errorf("user already exists, exiting now...")
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.args[0],
	})
	if err != nil {
		return err
	}
	if err = s.config.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Printf("User was created: %+v\n", user)

	return nil
}
