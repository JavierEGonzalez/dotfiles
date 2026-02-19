package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/types"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa")).
			Bold(true)
	infoTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))
	tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f5e0dc")).
			Background(lipgloss.Color("#45475a"))
	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6c7086"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086"))
)

const (
	TabChanges = iota
	TabAgent
	TabTicket
	TabPlan
	TabCount
)

type WorktreeModel struct {
	worktree  types.Worktree
	activeTab int
	err       error
}

func NewWorktreeModel(worktree types.Worktree) WorktreeModel {
	return WorktreeModel{
		worktree:  worktree,
		activeTab: TabChanges,
	}
}

func (m WorktreeModel) Init() tea.Cmd {
	return nil
}

func (m WorktreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			m.activeTab = TabChanges
			return m, nil
		case "a":
			m.activeTab = TabAgent
			return m, nil
		case "t":
			m.activeTab = TabTicket
			return m, nil
		case "p":
			m.activeTab = TabPlan
			return m, nil
		case "h", "left":
			m.activeTab = (m.activeTab - 1 + TabCount) % TabCount
			return m, nil
		case "l", "right":
			m.activeTab = (m.activeTab + 1) % TabCount
			return m, nil
		case "q", "esc":
			return m, func() tea.Msg { return WorktreeBackMsg{} }
		}
	}
	return m, nil
}

func (m WorktreeModel) View() string {
	var rows []string

	rows = append(rows, "")
	rows = append(rows, headerStyle.Render("  Worktree Details  "))
	rows = append(rows, "")
	rows = append(rows, infoTextStyle.Render(fmt.Sprintf("  Ticket: %s", m.worktree.Ticket)))
	rows = append(rows, infoTextStyle.Render(fmt.Sprintf("  Branch: %s", m.worktree.Branch)))
	rows = append(rows, infoTextStyle.Render(fmt.Sprintf("  Path:   %s", m.worktree.Path)))
	rows = append(rows, "")

	rows = append(rows, m.renderTabBar())
	rows = append(rows, "")
	rows = append(rows, m.renderTabContent())
	rows = append(rows, "")

	rows = append(rows, helpStyle.Render("  [g] Changes  [a] Agent  [t] Ticket  [p] Plan  |  [h/l] prev/next  [q/esc] back"))

	if m.err != nil {
		rows = append(rows, "")
		rows = append(rows, headerStyle.Render(fmt.Sprintf("  Error: %s", m.err.Error())))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m WorktreeModel) renderTabBar() string {
	tabs := []string{
		fmt.Sprintf("  [g] Changes"),
		fmt.Sprintf("  [a] Agent"),
		fmt.Sprintf("  [t] Ticket"),
		fmt.Sprintf("  [p] Plan"),
	}

	var renderedTabs []string
	for i, tab := range tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, tabActiveStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, tabInactiveStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m WorktreeModel) renderTabContent() string {
	tabContent := []string{
		"Changes tab - Use 'v' to view diff, 's' to stage, 'c' to commit",
		"Agent tab - Use 'o' to run OpenCode, 'r' to run Ralph",
		"Ticket tab - Use 'r' to refresh from Jira, 'e' to edit notes",
		"Plan tab - Use 'c' to generate plan, 'e' to edit, 'x' to execute",
	}

	content := tabContent[m.activeTab]
	return infoTextStyle.Render(fmt.Sprintf("  %s", content))
}

type WorktreeBackMsg struct{}
