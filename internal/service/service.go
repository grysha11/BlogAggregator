package service

import (
	"grysha11/BlogAggregator/internal/database"
	"grysha11/BlogAggregator/internal/config"
)

type State struct {
	DB 		*database.Queries
	Config	*config.Config
}

func New(db *database.Queries, config *config.Config) *State {
	return &State{
		DB: db,
		Config: config,
	}
}
