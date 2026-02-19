package main

import (
	"os"

	"github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return ""
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		os.Exit(1)
	}
}
