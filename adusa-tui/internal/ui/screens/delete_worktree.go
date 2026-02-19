package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/git"
	"github.com/javiergonzalez/adusa-tui/internal/types"
)

var (
	confirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f38ba8")).
			Bold(true)
	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f9e2af"))
	selectedButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa"))
)

type DeleteWorktreeModel struct {
	worktree types.Worktree
	choice   string // "y" or "n"
	err      error
}

func NewDeleteWorktreeModel(worktree types.Worktree) DeleteWorktreeModel {
	return DeleteWorktreeModel{
		worktree: worktree,
		choice:   "n",
	}
}

func (m DeleteWorktreeModel) Init() tea.Cmd {
	return nil
}

func (m DeleteWorktreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, func() tea.Msg { return DeleteWorktreeCancelledMsg{} }
		default:
			switch msg.String() {
			case "y":
				m.choice = "y"
				err := git.DeleteWorktree(m.worktree.Path)
				if err != nil {
					m.err = err
					return m, nil
				}
				return m, func() tea.Msg { return DeleteWorktreeDoneMsg{} }
			case "n":
				m.choice = "n"
				return m, func() tea.Msg { return DeleteWorktreeCancelledMsg{} }
			}
		}
	}
	return m, nil
}

func (m DeleteWorktreeModel) View() string {
	var rows []string

	rows = append(rows, "")
	rows = append(rows, confirmStyle.Render("  ⚠  Confirm Delete  ⚠  "))
	rows = append(rows, "")
	rows = append(rows, infoStyle.Render(fmt.Sprintf("  Worktree: %s", m.worktree.Path)))
	rows = append(rows, infoStyle.Render(fmt.Sprintf("  Ticket:   %s", m.worktree.Ticket)))
	rows = append(rows, infoStyle.Render(fmt.Sprintf("  Branch:  %s", m.worktree.Branch)))
	rows = append(rows, "")
	rows = append(rows, confirmStyle.Render("  This will remove git worktree and delete directory"))
	rows = append(rows, infoStyle.Render("  Branch will NOT be deleted"))
	rows = append(rows, "")

	if m.err != nil {
		rows = append(rows, "")
		rows = append(rows, confirmStyle.Render(fmt.Sprintf("  Error: %s", m.err.Error())))
		rows = append(rows, "")
	}

	yesBtn := buttonStyle.Render("[y] Yes, delete")
	noBtn := buttonStyle.Render("[n] No, cancel")

	if m.choice == "y" {
		yesBtn = selectedButton.Render("[y] Yes, delete")
	} else {
		noBtn = selectedButton.Render("[n] No, cancel")
	}

	rows = append(rows, fmt.Sprintf("  %s  %s", yesBtn, noBtn))
	rows = append(rows, "")

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

type DeleteWorktreeCancelledMsg struct{}

type DeleteWorktreeDoneMsg struct{}
