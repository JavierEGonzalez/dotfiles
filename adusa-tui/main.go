package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/ui/screens"
)

type model struct {
	worktreesScreen      screens.WorktreesModel
	createWorktreeScreen screens.CreateWorktreeModel
	deleteWorktreeScreen screens.DeleteWorktreeModel
	worktreeScreen       screens.WorktreeModel
	view                 string
	width                int
	height               int
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.view {
	case "create":
		return m.updateCreateWorktree(msg)
	case "delete":
		return m.updateDeleteWorktree(msg)
	case "worktree":
		return m.updateWorktree(msg)
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
		m.view = "worktree"
		m.worktreeScreen = screens.NewWorktreeModel(msg.Worktree)
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

func (m model) updateWorktree(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.WorktreeBackMsg:
		m.view = ""
		return m, nil
	default:
		newScreen, cmd := m.worktreeScreen.Update(msg)
		m.worktreeScreen = newScreen.(screens.WorktreeModel)
		return m, cmd
	}
}

func (m model) View() string {
	var content string
	switch m.view {
	case "create":
		content = m.createWorktreeScreen.View()
	case "delete":
		content = m.deleteWorktreeScreen.View()
	case "worktree":
		content = m.worktreeScreen.View()
	default:
		// For the main worktrees view, center it if there's room
		if m.height > 0 && m.width > 0 {
			content = centerContent(m.worktreesScreen.View(), m.width, m.height)
		} else {
			content = m.worktreesScreen.View()
		}
		return content
	}
	return content
}

func centerContent(content string, width, height int) string {
	// Use lipgloss to center the content
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	return style.Render(content)
}

func main() {
	if _, err := tea.NewProgram(model{
		worktreesScreen: screens.NewWorktreesModel(),
	}).Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
