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
		return fmt.Errorf("incorrect amount of arguments in command call: %v\nUsage: login <argument>", cmd.name)
	}

	checkExist, err := s.db.GetUserByName(context.Background(), cmd.args[0])
	if checkExist.ID == uuid.Nil {
		return fmt.Errorf("user don't exist")
	}
	if err != nil {
		return err
	}

	if err := s.config.SetUser(cmd.args[0]); err != nil {
		return err
	}

	fmt.Printf("Logged as: %v\n", s.config.CurrentUsername)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: %v\nUsage: register <argument>", cmd.name)
	}

	checkDup, err := s.db.GetUserByName(context.Background(), cmd.args[0])
	if err == nil && checkDup.ID != uuid.Nil {
		return fmt.Errorf("user already exists, exiting now...")
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
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

func handlerReset(s *state, cmd command) error {
	//i will not add checker for args it will execute anyway

	err := s.db.DeleteUsers(context.Background())
	err = s.db.DeleteFeeds(context.Background())
	err = s.db.DeleteFeedFollows(context.Background())

	return err
}

func handlerUsers(s *state, cmd command) error {
	//same here with args checker

	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Printf("There are no users yet!")
		return nil
	}

	for _, user := range users {
		fmt.Printf("* %v", user.Name)
		if user.Name == s.config.CurrentUsername {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: agg <time_between_reqs/1h,1m,1s>", cmd.name)
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("error during parsing time: %v\nUsage: agg <time_between_reqs/1h,1m,1s>", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

//TODO add checker for dups of feeds

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: addfeed <feed_name> <feed_url>", cmd.name)
	}

	checkDup, err := s.db.GetFeedByURL(context.Background(), cmd.args[1])
	if err == nil && checkDup.ID != uuid.Nil {
		return fmt.Errorf("feed already exists, exiting now...")
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:	uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: cmd.args[0],
		Url: cmd.args[1],
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
		FeedID: feed.ID,
	})

	fmt.Printf("Feed was created: %+v\n", follow)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Printf("There are no feeds yet!\n")
		return nil
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("*\t%v\n\t %v\n\t %v\n", feed.Name, feed.Url, user.Name)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: follow <feed_url>", cmd.name)
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("Feed doesn't exist: %v", err)
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Feed:\t%v\nUser:\t%v\n", follow.FeedName, follow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Printf("You don't have any feeds yet!\n")
		return nil
	}

	fmt.Printf("Feeds which %v follows:\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("\t* %v\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: unfollow <feed_url>", cmd.name)
	}

	err := s.db.DeleteFeedFollowByUrl(context.Background(), database.DeleteFeedFollowByUrlParams{
		UserID: user.ID,
		Url: cmd.args[0],
	})
	if err != nil {
		return err
	}

	fmt.Printf("Feed was unfollowed: %v\n", cmd.args[0])
	return nil
}
