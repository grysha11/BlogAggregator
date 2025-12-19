package main

import (
	"fmt"
	"grysha11/BlogAggregator/internal/config"
)

type state struct {
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

	if err := s.config.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Printf("Username have been set to %v\n", s.config.CurrentUsername)
	return nil
}



