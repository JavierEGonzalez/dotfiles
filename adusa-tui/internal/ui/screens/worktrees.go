package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/git"
	"github.com/javiergonzalez/adusa-tui/internal/types"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffd700")).
			Bold(true)
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa"))
	statusClean = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1"))
	statusDirty = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
)

type WorktreesModel struct {
	list        list.Model
	worktrees   []types.Worktree
	dirtyStatus map[string]bool
	selectedIdx int
}

func NewWorktreesModel() WorktreesModel {
	wt, _ := git.ListWorktrees()

	dirtyStatus := make(map[string]bool)
	for i := range wt {
		status, _ := git.GetStatus(wt[i].Path)
		dirtyStatus[wt[i].Path] = len(status) > 0
	}

	items := make([]list.Item, len(wt))
	for i, w := range wt {
		items[i] = w
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Worktrees"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return WorktreesModel{
		list:        l,
		worktrees:   wt,
		dirtyStatus: dirtyStatus,
	}
}

func (m WorktreesModel) Init() tea.Cmd {
	return nil
}

func (m WorktreesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if len(m.worktrees) > 0 {
				m.selectedIdx = m.list.Index()
				return m, func() tea.Msg { return WorktreeSelectedMsg{Worktree: m.worktrees[m.selectedIdx]} }
			}
		default:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "n":
				return m, func() tea.Msg { return CreateWorktreeMsg{} }
			case "d":
				if len(m.worktrees) > 0 {
					m.selectedIdx = m.list.Index()
					return m, func() tea.Msg { return DeleteWorktreeMsg{Worktree: m.worktrees[m.selectedIdx]} }
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m WorktreesModel) View() string {
	if len(m.worktrees) == 0 {
		return titleStyle.Render("No worktrees found. Press 'n' to create one.")
	}

	selected := m.list.Index()
	var rows []string
	rows = append(rows, titleStyle.Render("All Worktrees"))
	rows = append(rows, "")
	rows = append(rows, fmt.Sprintf("  %-10s %-30s %s", "TICKET", "BRANCH", "STATUS"))

	for i, wt := range m.worktrees {
		ticketStr := wt.Ticket
		if ticketStr == "" {
			ticketStr = "(none)"
		}
		branchStr := wt.Branch
		if len(branchStr) > 30 {
			branchStr = branchStr[:27] + "..."
		}

		statusStr := statusClean.Render("clean")
		if m.dirtyStatus[wt.Path] {
			statusStr = statusDirty.Render("dirty")
		}

		prefix := "  "
		if i == selected {
			prefix = "► "
			ticketStr = selectedStyle.Render(ticketStr)
			branchStr = selectedStyle.Render(branchStr)
			statusStr = selectedStyle.Render(statusStr[lipgloss.Width(statusStr)-len(statusStr):])
		}

		rows = append(rows, fmt.Sprintf("%s%-10s %-30s %s", prefix, ticketStr, branchStr, statusStr))
	}

	rows = append(rows, "")
	rows = append(rows, "  ↑↓/j/k: navigate  Enter: select  n: new  d: delete  q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m WorktreesModel) Refresh() (WorktreesModel, error) {
	wt, err := git.ListWorktrees()
	if err != nil {
		return m, err
	}

	dirtyStatus := make(map[string]bool)
	for i := range wt {
		status, _ := git.GetStatus(wt[i].Path)
		dirtyStatus[wt[i].Path] = len(status) > 0
	}

	items := make([]list.Item, len(wt))
	for i, w := range wt {
		items[i] = w
	}

	m.list.SetItems(items)
	m.worktrees = wt
	m.dirtyStatus = dirtyStatus

	return m, nil
}

func (m WorktreesModel) GetSelectedWorktree() *types.Worktree {
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.worktrees) {
		return &m.worktrees[m.selectedIdx]
	}
	return nil
}

type WorktreeSelectedMsg struct {
	Worktree types.Worktree
}

type CreateWorktreeMsg struct{}

type DeleteWorktreeMsg struct {
	Worktree types.Worktree
}
