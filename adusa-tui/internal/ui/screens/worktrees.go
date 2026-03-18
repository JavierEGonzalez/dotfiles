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
	showingHelp bool

	list        list.Model
	worktrees   []types.Worktree
	dirtyStatus map[string]bool
	selectedIdx int
	width       int
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

func (m *WorktreesModel) SetWidth(w int) {
	m.width = w
}

func (m WorktreesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.showingHelp {
				m.showingHelp = false
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyEnter:
			if len(m.worktrees) > 0 {
				m.selectedIdx = m.list.Index()
				return m, func() tea.Msg { return WorktreeSelectedMsg{Worktree: m.worktrees[m.selectedIdx]} }
			}
		default:
			switch msg.String() {
			case "?":
				m.showingHelp = !m.showingHelp
				return m, nil
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
	if m.showingHelp {
		return m.renderHelp()
	}
	if len(m.worktrees) == 0 {
		return titleStyle.Render("No worktrees found. Press 'n' to create one.")
	}

	selected := m.list.Index()
	var rows []string
	rows = append(rows, titleStyle.Render("All Worktrees"))
	rows = append(rows, "")

	// Fixed columns: prefix(2) + ticket(12) + status(8)
	// Branch gets the remaining width
	ticketWidth := 12
	statusWidth := 8
	fixedWidth := 2 + ticketWidth + statusWidth + 2 // 2 for spacing between columns
	branchWidth := m.width - fixedWidth
	if branchWidth < 20 {
		branchWidth = 20
	}

	rows = append(rows, fmt.Sprintf("  %-*s %-*s %s", ticketWidth, "TICKET", branchWidth, "BRANCH", "STATUS"))

	for i, wt := range m.worktrees {
		ticketStr := wt.Ticket
		if ticketStr == "" {
			ticketStr = "(none)"
		}
		branchStr := wt.Branch
		statusLabel := "clean"
		isDirty := m.dirtyStatus[wt.Path]
		if isDirty {
			statusLabel = "dirty"
		}

		// Pad plain text first, then apply styling so ANSI codes don't break alignment
		paddedTicket := fmt.Sprintf("%-*s", ticketWidth, ticketStr)
		paddedBranch := fmt.Sprintf("%-*s", branchWidth, branchStr)

		prefix := "  "
		if i == selected {
			prefix = "► "
			paddedTicket = selectedStyle.Render(paddedTicket)
			paddedBranch = selectedStyle.Render(paddedBranch)
			statusLabel = selectedStyle.Render(statusLabel)
		} else if isDirty {
			statusLabel = statusDirty.Render(statusLabel)
		} else {
			statusLabel = statusClean.Render(statusLabel)
		}

		rows = append(rows, fmt.Sprintf("%s%s %s %s", prefix, paddedTicket, paddedBranch, statusLabel))
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

func (m WorktreesModel) FindByTicket(ticket string) *types.Worktree {
	for i, wt := range m.worktrees {
		if wt.Ticket == ticket {
			m.selectedIdx = i
			return &m.worktrees[i]
		}
	}
	return nil
}

func (m WorktreesModel) FindByPath(path string) *types.Worktree {
	for i, wt := range m.worktrees {
		if wt.Path == path {
			m.selectedIdx = i
			return &m.worktrees[i]
		}
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

func (m WorktreesModel) renderHelp() string {
	var lines []string
	lines = append(lines, titleStyle.Render("Keyboard Shortcuts - All Worktrees"))
	lines = append(lines, "")
	lines = append(lines, "  ↑/k:    Move up")
	lines = append(lines, "  ↓/j:    Move down")
	lines = append(lines, "  Enter:  Select worktree")
	lines = append(lines, "  n:      Create new worktree")
	lines = append(lines, "  d:      Delete selected worktree")
	lines = append(lines, "  ?:      Toggle help")
	lines = append(lines, "  q/Esc:  Quit")
	lines = append(lines, "")
	lines = append(lines, "Press ? or Esc to return")
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
