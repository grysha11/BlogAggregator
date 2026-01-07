package cli

import (
	"context"
	"database/sql"
	"fmt"
	"grysha11/BlogAggregator/internal/database"
	"grysha11/BlogAggregator/internal/service"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Command struct {
	Name		string
	Args		[]string
}

type Commands struct {
	knownCommands	map[string]func(*service.State, Command) error
}

func (c *Commands) Run(s *service.State, cmd Command) error {
	var err error

	if fun, ok := c.knownCommands[cmd.Name]; ok {
		err = fun(s, cmd)
	} else {
		err = fmt.Errorf("unknown command")
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) Register(name string, f func(*service.State, Command) error) {
	if c.knownCommands == nil {
		c.knownCommands = make(map[string]func(*service.State, Command) error)
	}
	c.knownCommands[name] = f
}

func HandlerLogin(s *service.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: %v\nUsage: login <argument>", cmd.Name)
	}

	checkExist, err := s.DB.GetUserByName(context.Background(), cmd.Args[0])
	if checkExist.ID == uuid.Nil {
		return fmt.Errorf("user don't exist")
	}
	if err != nil {
		return err
	}

	if err := s.Config.SetUser(cmd.Args[0]); err != nil {
		return err
	}

	fmt.Printf("Logged as: %v\n", s.Config.CurrentUsername)
	return nil
}

func HandlerRegister(s *service.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: %v\nUsage: register <argument>", cmd.Name)
	}

	checkDup, err := s.DB.GetUserByName(context.Background(), cmd.Args[0])
	if err == nil && checkDup.ID != uuid.Nil {
		return fmt.Errorf("user already exists, exiting now...")
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	user, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: cmd.Args[0],
	})
	if err != nil {
		return err
	}
	if err = s.Config.SetUser(cmd.Args[0]); err != nil {
		return err
	}

	fmt.Printf("User was created: %+v\n", user)

	return nil
}

func HandlerReset(s *service.State, cmd Command) error {
	//i will not add checker for args it will execute anyway

	err := s.DB.DeleteUsers(context.Background())
	err = s.DB.DeleteFeeds(context.Background())
	err = s.DB.DeleteFeedFollows(context.Background())

	return err
}

func HandlerUsers(s *service.State, cmd Command) error {
	//same here with args checker

	users, err := s.DB.GetAllUsers(context.Background())
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Printf("There are no users yet!")
		return nil
	}

	for _, user := range users {
		fmt.Printf("* %v", user.Name)
		if user.Name == s.Config.CurrentUsername {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}

	return nil
}

func HandlerAgg(s *service.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: agg <time_between_reqs/1h,1m,1s>", cmd.Name)
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error during parsing time: %v\nUsage: agg <time_between_reqs/1h,1m,1s>", err)
	}

	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	
	for ; ; <-ticker.C {
		service.ScrapeFeeds(s)
	}
}

//TODO add checker for dups of feeds

func HandlerAddFeed(s *service.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: addfeed <feed_name> <feed_url>", cmd.Name)
	}

	checkDup, err := s.DB.GetFeedByURL(context.Background(), cmd.Args[1])
	if err == nil && checkDup.ID != uuid.Nil {
		return fmt.Errorf("feed already exists, exiting now...")
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	feed, err := s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:	uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	follow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
		FeedID: feed.ID,
	})

	fmt.Printf("Feed was created: %+v\n", follow)
	return nil
}

func HandlerFeeds(s *service.State, cmd Command) error {
	feeds, err := s.DB.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		fmt.Printf("There are no feeds yet!\n")
		return nil
	}

	for _, feed := range feeds {
		user, err := s.DB.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("*\t%v\n\t %v\n\t %v\n", feed.Name, feed.Url, user.Name)
	}

	return nil
}

func HandlerFollow(s *service.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: follow <feed_url>", cmd.Name)
	}

	feed, err := s.DB.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Feed doesn't exist: %v", err)
	}

	follow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
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

func HandlerFollowing(s *service.State, cmd Command, user database.User) error {
	feeds, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
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

func HandlerUnfollow(s *service.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: unfollow <feed_url>", cmd.Name)
	}

	err := s.DB.DeleteFeedFollowByUrl(context.Background(), database.DeleteFeedFollowByUrlParams{
		UserID: user.ID,
		Url: cmd.Args[0],
	})
	if err != nil {
		return err
	}

	fmt.Printf("Feed was unfollowed: %v\n", cmd.Args[0])
	return nil
}

func HandlerBrowse(s *service.State, cmd Command, user database.User) error {
	var limit int32
	limit = 2

	if len(cmd.Args) > 1 {
		return fmt.Errorf("incorrect amount of arguments in command call: <%v>\nUsage: browse *Optional:<limit>", cmd.Name)
	}

	if len(cmd.Args) == 1 {
		if manualLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = int32(manualLimit)
		} else {
			return fmt.Errorf("invalid limit provided: %v", cmd.Args[0])
		}
	}

	posts, err := s.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: limit,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Found %v posts for user %v:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("    Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}
	
	return nil
}
