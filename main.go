package main

import (
	"fmt"
	"os"

	"grysha11/BlogAggregator/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Too little amount of arguments :(")
		os.Exit(1)
	}
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error while reading config file: %v\n", err)
		os.Exit(1)
	}

	s := state{}
	s.config = &cfg

	cmds := commands{}
	cmds.register("login", handlerLogin)

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err = cmds.run(&s, cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}