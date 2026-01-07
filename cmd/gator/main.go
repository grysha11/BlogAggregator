package main

import (
	"database/sql"
	"fmt"
	"os"

	"grysha11/BlogAggregator/internal/config"
	"grysha11/BlogAggregator/internal/database"
	"grysha11/BlogAggregator/internal/service"
	"grysha11/BlogAggregator/internal/cli"

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

	s := service.New(database.New(db), &cfg)

	cmds := cli.Commands{}
	cmds.Register("login", cli.HandlerLogin)
	cmds.Register("register", cli.HandlerRegister)
	cmds.Register("reset", cli.HandlerReset)
	cmds.Register("users", cli.HandlerUsers)
	cmds.Register("agg", cli.HandlerAgg)
	cmds.Register("feeds", cli.HandlerFeeds)
	cmds.Register("addfeed", cli.MiddlewareLoggedIn(cli.HandlerAddFeed))
	cmds.Register("follow", cli.MiddlewareLoggedIn(cli.HandlerFollow))
	cmds.Register("following", cli.MiddlewareLoggedIn(cli.HandlerFollowing))
	cmds.Register("unfollow", cli.MiddlewareLoggedIn(cli.HandlerUnfollow))
	cmds.Register("browse", cli.MiddlewareLoggedIn(cli.HandlerBrowse))

	cmd := cli.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err = cmds.Run(s, cmd); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}