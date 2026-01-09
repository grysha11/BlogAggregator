package main

import (
	"database/sql"
	"fmt"
	"os"

	"grysha11/BlogAggregator/internal/config"
	"grysha11/BlogAggregator/internal/database"
	"grysha11/BlogAggregator/internal/service"
	"grysha11/BlogAggregator/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		fmt.Printf("Error while connecting to database: %v\n", err)
		os.Exit(1)
	}

	s := service.New(database.New(db), &cfg)

	p := tea.NewProgram(ui.InitialModel(s))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error occured: %v\n", err)
		os.Exit(1)
	}
}