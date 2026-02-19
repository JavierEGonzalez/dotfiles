package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/javiergonzalez/adusa-tui/internal/ui/screens"
)

type model struct {
	worktreesScreen      screens.WorktreesModel
	createWorktreeScreen screens.CreateWorktreeModel
	deleteWorktreeScreen screens.DeleteWorktreeModel
	view                 string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.view {
	case "create":
		return m.updateCreateWorktree(msg)
	case "delete":
		return m.updateDeleteWorktree(msg)
	default:
		return m.updateWorktrees(msg)
	}
}

func (m model) updateWorktrees(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case screens.WorktreeSelectedMsg:
		fmt.Printf("Selected worktree: %s\n", msg.Worktree.Path)
		return m, nil
	case screens.CreateWorktreeMsg:
		m.view = "create"
		m.createWorktreeScreen = screens.NewCreateWorktreeModel()
		return m, nil
	case screens.DeleteWorktreeMsg:
		m.view = "delete"
		m.deleteWorktreeScreen = screens.NewDeleteWorktreeModel(msg.Worktree)
		return m, nil
	}

	newScreen, cmd := m.worktreesScreen.Update(msg)
	m.worktreesScreen = newScreen.(screens.WorktreesModel)
	return m, cmd
}

func (m model) updateCreateWorktree(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.CreateWorktreeDoneMsg:
		m.view = ""
		wt, _ := m.worktreesScreen.Refresh()
		m.worktreesScreen = wt
		return m, nil
	default:
		newScreen, cmd := m.createWorktreeScreen.Update(msg)
		m.createWorktreeScreen = newScreen.(screens.CreateWorktreeModel)
		return m, cmd
	}
}

func (m model) updateDeleteWorktree(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case screens.DeleteWorktreeDoneMsg:
		m.view = ""
		wt, _ := m.worktreesScreen.Refresh()
		m.worktreesScreen = wt
		return m, nil
	case screens.DeleteWorktreeCancelledMsg:
		m.view = ""
		return m, nil
	default:
		newScreen, cmd := m.deleteWorktreeScreen.Update(msg)
		m.deleteWorktreeScreen = newScreen.(screens.DeleteWorktreeModel)
		return m, cmd
	}
}

func (m model) View() string {
	switch m.view {
	case "create":
		return m.createWorktreeScreen.View()
	case "delete":
		return m.deleteWorktreeScreen.View()
	default:
		return m.worktreesScreen.View()
	}
}

func main() {
	if _, err := tea.NewProgram(model{
		worktreesScreen: screens.NewWorktreesModel(),
	}).Run(); err != nil {
		os.Exit(1)
	}
}
