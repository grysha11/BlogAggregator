package main

import (
	"database/sql"
	"fmt"
	"os"

	"grysha11/BlogAggregator/internal/config"
	"grysha11/BlogAggregator/internal/database"

	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		fmt.Printf("Error while connecting to database: %v\n", err)
		os.Exit(1)
	}

	s := state{}
	s.db = database.New(db)
	s.config = &cfg

	cmds := commands{}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err = cmds.run(&s, cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}