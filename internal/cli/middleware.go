package cli

import (
	"context"
	"grysha11/BlogAggregator/internal/database"
    "grysha11/BlogAggregator/internal/service"
)

func MiddlewareLoggedIn(handler func(s *service.State, cmd Command, user database.User) error) func(*service.State, Command) error {
    return func(s *service.State, cmd Command) error {
        user, err := s.DB.GetUserByName(context.Background(), s.Config.CurrentUsername)
        if err != nil {
            return err
        }

        return handler(s, cmd, user)
    }
}