package screens

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/git"
	"github.com/javiergonzalez/adusa-tui/internal/jira"
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
	statusModifiedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f9e2af"))
	statusAddedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a6e3a1"))
	statusUntrackedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9399b2"))
	statusDeletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f38ba8"))
	diffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa"))
	ticketLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#89b4fa"))
	ticketValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#cdd6f4"))
)

const (
	TabChanges = iota
	TabAgent
	TabTicket
	TabPlan
	TabCount
)

type WorktreeModel struct {
	worktree     types.Worktree
	activeTab    int
	err          error
	status       []git.FileStatus
	diff         string
	showingDiff  bool
	diffScroll   int
	commitMode   bool
	commitMsg    string
	isLoading    bool
	loadingMsg   string
	ticket       *types.TicketInfo
	ticketScroll int
}

func NewWorktreeModel(worktree types.Worktree) WorktreeModel {
	m := WorktreeModel{
		worktree:  worktree,
		activeTab: TabChanges,
	}
	m.loadStatus()
	return m
}

func (m WorktreeModel) Init() tea.Cmd {
	return nil
}

func (m *WorktreeModel) loadStatus() {
	m.isLoading = true
	m.loadingMsg = "Loading git status..."
	status, err := git.GetStatus(m.worktree.Path)
	m.isLoading = false
	if err != nil {
		m.err = err
		return
	}
	m.status = status
}

func (m *WorktreeModel) loadDiff() {
	m.isLoading = true
	m.loadingMsg = "Loading diff..."
	diff, err := git.GetDiff(m.worktree.Path)
	m.isLoading = false
	if err != nil {
		m.err = err
		return
	}
	m.diff = diff
	m.diffScroll = 0
}

func (m *WorktreeModel) loadTicket() {
	m.isLoading = true
	m.loadingMsg = "Loading ticket..."
	ticket, err := jira.LoadTicketCache(m.worktree.Ticket)
	if err != nil {
		m.err = err
		m.isLoading = false
		return
	}
	if ticket == nil {
		m.fetchTicket()
		return
	}
	m.isLoading = false
	m.ticket = ticket
	m.ticketScroll = 0
}

func (m *WorktreeModel) fetchTicket() {
	m.isLoading = true
	m.loadingMsg = "Fetching from Jira..."
	ticket, err := jira.FetchTicket(m.worktree.Ticket)
	m.isLoading = false
	if err != nil {
		m.err = err
		return
	}
	if err := jira.SaveTicketCache(ticket); err != nil {
		m.err = err
		return
	}
	m.ticket = ticket
	m.ticketScroll = 0
}

func (m *WorktreeModel) editTicket() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, m.getTicketCachePath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	m.loadTicket()
}

func (m *WorktreeModel) appendToTicketNotes() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cachePath := m.getTicketCachePath()
	cmd := exec.Command(editor, cachePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	m.loadTicket()
}

func (m *WorktreeModel) getTicketCachePath() string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/.scratch/tickets/%s.md", homeDir, m.worktree.Ticket)
}

func (m WorktreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showingDiff {
		return m.updateDiffView(msg)
	}
	if m.commitMode {
		return m.updateCommitMode(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			m.activeTab = TabChanges
			return m, nil
		case "a":
			if m.activeTab == TabTicket {
				m.appendToTicketNotes()
				return m, nil
			}
			m.activeTab = TabAgent
			return m, nil
		case "t":
			m.activeTab = TabTicket
			m.loadTicket()
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
		case "r":
			if m.activeTab == TabChanges {
				m.loadStatus()
			} else if m.activeTab == TabTicket {
				m.fetchTicket()
			}
			return m, nil
		case "v":
			if m.activeTab == TabChanges {
				m.loadDiff()
				m.showingDiff = true
			}
			return m, nil
		case "s":
			if m.activeTab == TabChanges {
				m.isLoading = true
				m.loadingMsg = "Staging all files..."
				err := git.StageAll(m.worktree.Path)
				m.isLoading = false
				if err != nil {
					m.err = err
				} else {
					m.loadStatus()
				}
			}
			return m, nil
		case "c":
			if m.activeTab == TabChanges && len(m.status) > 0 {
				m.commitMode = true
				m.commitMsg = fmt.Sprintf("[%s]: ", m.worktree.Ticket)
			}
			return m, nil
		case "e":
			if m.activeTab == TabTicket {
				m.editTicket()
			}
			return m, nil
		case "j", "down":
			if m.activeTab == TabTicket {
				m.ticketScroll++
			}
			return m, nil
		case "k", "up":
			if m.activeTab == TabTicket && m.ticketScroll > 0 {
				m.ticketScroll--
			}
			return m, nil
		}
	}
	return m, nil
}

func (m WorktreeModel) updateDiffView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.showingDiff = false
			m.diff = ""
			m.diffScroll = 0
			return m, nil
		case "j", "down":
			m.diffScroll++
			return m, nil
		case "k", "up":
			if m.diffScroll > 0 {
				m.diffScroll--
			}
			return m, nil
		}
	}
	return m, nil
}

func (m WorktreeModel) updateCommitMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			msg := strings.TrimPrefix(m.commitMsg, fmt.Sprintf("[%s]: ", m.worktree.Ticket))
			if msg == "" {
				m.err = fmt.Errorf("commit message cannot be empty")
				m.commitMode = false
				return m, nil
			}
			m.isLoading = true
			m.loadingMsg = "Committing..."
			err := git.Commit(m.worktree.Path, m.worktree.Ticket, msg)
			m.isLoading = false
			if err != nil {
				m.err = err
			} else {
				m.loadStatus()
			}
			m.commitMode = false
			m.commitMsg = ""
			return m, nil
		case tea.KeyEscape:
			m.commitMode = false
			m.commitMsg = ""
			return m, nil
		case tea.KeyBackspace:
			if len(m.commitMsg) > 0 {
				m.commitMsg = m.commitMsg[:len(m.commitMsg)-1]
			}
			return m, nil
		}
		if len(msg.String()) == 1 {
			m.commitMsg += msg.String()
		}
	}
	return m, nil
}

func (m WorktreeModel) View() string {
	if m.showingDiff {
		return m.renderDiffView()
	}

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

	if m.activeTab == TabChanges {
		rows = append(rows, helpStyle.Render("  [v] view diff  [s] stage  [c] commit  [r] refresh  |  [g/a/t/p] tabs  [h/l] prev/next  [q/esc] back"))
	} else {
		rows = append(rows, helpStyle.Render("  [g] Changes  [a] Agent  [t] Ticket  [p] Plan  |  [h/l] prev/next  [q/esc] back"))
	}

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
	if m.isLoading {
		return infoTextStyle.Render(fmt.Sprintf("  %s", m.loadingMsg))
	}

	switch m.activeTab {
	case TabChanges:
		return m.renderChangesTab()
	case TabAgent:
		return infoTextStyle.Render("  Agent tab - Use 'o' to run OpenCode, 'r' to run Ralph")
	case TabTicket:
		return m.renderTicketTab()
	case TabPlan:
		return infoTextStyle.Render("  Plan tab - Use 'c' to generate plan, 'e' to edit, 'x' to execute")
	}
	return ""
}

func (m WorktreeModel) renderChangesTab() string {
	if m.commitMode {
		return promptStyle.Render(fmt.Sprintf("  Commit message: %s", m.commitMsg))
	}

	if len(m.status) == 0 {
		return infoTextStyle.Render("  No changes detected")
	}

	var lines []string
	lines = append(lines, infoTextStyle.Render("  Status:"))

	for _, s := range m.status {
		var statusStr string
		switch s.Status {
		case "M":
			statusStr = statusModifiedStyle.Render("M")
		case "A":
			statusStr = statusAddedStyle.Render("A")
		case "D":
			statusStr = statusDeletedStyle.Render("D")
		case "?":
			statusStr = statusUntrackedStyle.Render("?")
		default:
			statusStr = s.Status
		}
		lines = append(lines, infoTextStyle.Render(fmt.Sprintf("    %s  %s", statusStr, s.Path)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m WorktreeModel) renderDiffView() string {
	var rows []string

	rows = append(rows, "")
	rows = append(rows, headerStyle.Render("  Diff View  "))
	rows = append(rows, "")

	if m.diff == "" {
		rows = append(rows, infoTextStyle.Render("  No diff to display"))
	} else {
		diffLines := strings.Split(m.diff, "\n")
		start := m.diffScroll
		end := m.diffScroll + 20
		if end > len(diffLines) {
			end = len(diffLines)
		}
		if start > 0 {
			rows = append(rows, infoTextStyle.Render("  ... (scroll up with k)"))
		}
		for i := start; i < end; i++ {
			rows = append(rows, diffStyle.Render("  "+diffLines[i]))
		}
		if end < len(diffLines) {
			rows = append(rows, infoTextStyle.Render("  ... (scroll down with j)"))
		}
	}

	rows = append(rows, "")
	rows = append(rows, helpStyle.Render("  [j/k] scroll  [esc/q] back"))

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m WorktreeModel) renderTicketTab() string {
	if m.ticket == nil {
		if m.err != nil {
			return infoTextStyle.Render("  Error loading ticket: " + m.err.Error())
		}
		return infoTextStyle.Render("  No ticket information available. Press 'r' to fetch.")
	}

	var lines []string

	lines = append(lines, ticketLabelStyle.Render("  Summary: ")+ticketValueStyle.Render(m.ticket.Summary))
	lines = append(lines, ticketLabelStyle.Render("  ─────────────────────────────────────────────────"))
	lines = append(lines, fmt.Sprintf("  %s %s  │  %s %s  │  %s %s",
		ticketLabelStyle.Render("Status:"), ticketValueStyle.Render(m.ticket.Status),
		ticketLabelStyle.Render("Assignee:"), ticketValueStyle.Render(m.ticket.Assignee),
		ticketLabelStyle.Render("Priority:"), ticketValueStyle.Render(m.ticket.Priority),
	))
	lines = append(lines, "")
	lines = append(lines, ticketLabelStyle.Render("  Description:"))
	lines = append(lines, ticketLabelStyle.Render("  ─────────────────────────────────────────────────"))

	descLines := strings.Split(m.ticket.Description, "\n")
	start := m.ticketScroll
	end := start + 10
	if end > len(descLines) {
		end = len(descLines)
	}

	if start > 0 {
		lines = append(lines, infoTextStyle.Render("  ... (scroll up with k)"))
	}
	for i := start; i < end; i++ {
		lines = append(lines, ticketValueStyle.Render("  "+descLines[i]))
	}
	if end < len(descLines) {
		lines = append(lines, infoTextStyle.Render("  ... (scroll down with j)"))
	}

	lines = append(lines, ticketLabelStyle.Render("  ─────────────────────────────────────────────────"))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  [r] Refetch from Jira  [a] Append Notes  [e] Edit File"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type WorktreeBackMsg struct{}
