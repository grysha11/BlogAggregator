package ui

import (
	"grysha11/BlogAggregator/internal/service"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	s	*service.State
}

func InitialModel(s *service.State) *Model {
	return &Model{
		s: s,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit 
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := "Welcome to Gator TUI\n\n"

	s += "Connected as: " + m.s.Config.CurrentUsername + "\n\n"

	s += "Press 'q' to quit."

	return s
}