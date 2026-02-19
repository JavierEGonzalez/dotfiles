package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/javiergonzalez/adusa-tui/internal/ui/screens"
)

type model struct {
	worktreesScreen screens.WorktreesModel
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case screens.WorktreeSelectedMsg:
		fmt.Printf("Selected worktree: %s\n", msg.Worktree.Path)
		return m, nil
	case screens.CreateWorktreeMsg:
		fmt.Println("Create worktree requested")
		return m, nil
	case screens.DeleteWorktreeMsg:
		fmt.Printf("Delete worktree: %s\n", msg.Worktree.Path)
		return m, nil
	}

	newScreen, cmd := m.worktreesScreen.Update(msg)
	m.worktreesScreen = newScreen.(screens.WorktreesModel)
	return m, cmd
}

func (m model) View() string {
	return m.worktreesScreen.View()
}

func main() {
	if _, err := tea.NewProgram(model{
		worktreesScreen: screens.NewWorktreesModel(),
	}).Run(); err != nil {
		os.Exit(1)
	}
}
