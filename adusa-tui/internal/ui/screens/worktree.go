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

// Available AI models for agent selection
const (
	ModelClaudeSonnet    = "claude-3-5-sonnet-20241022"
	ModelClaudeOpus      = "claude-3-opus-20250219"
	ModelClaude3Haiku    = "claude-3-haiku-20240307"
	ModelGPT4            = "gpt-4"
	ModelGPT4Turbo       = "gpt-4-turbo"
	ModelGPT35Turbo      = "gpt-3.5-turbo"
	ModelGitHubCopilot   = "github-copilot"
	ModelLocalOpenCodeAI = "local-opencode"
)

var AvailableModels = []string{
	ModelGitHubCopilot,
	ModelClaudeSonnet,
	ModelClaudeOpus,
	ModelClaude3Haiku,
	ModelGPT4,
	ModelGPT4Turbo,
	ModelGPT35Turbo,
	ModelLocalOpenCodeAI,
}

type WorktreeModel struct {
	planContent      string
	planScroll       int
	planExists       bool
	confirmPlan      bool
	confirmExecute   bool
	worktree         types.Worktree
	activeTab        int
	err              error
	status           []git.FileStatus
	diff             string
	showingDiff      bool
	diffScroll       int
	commitMode       bool
	commitMsg        string
	isLoading        bool
	loadingMsg       string
	ticket           *types.TicketInfo
	ticketScroll     int
	agentRunning     bool
	agentDone        bool
	promptModel      bool
	promptIter       bool
	agentModel       string
	agentIter        string
	showingHelp      bool
	modelSelectorIdx int
	selectingModel   bool
}

func NewWorktreeModel(worktree types.Worktree) WorktreeModel {
	m := WorktreeModel{
		showingHelp:      false,
		worktree:         worktree,
		activeTab:        TabChanges,
		agentModel:       ModelClaudeSonnet,
		agentIter:        "3",
		modelSelectorIdx: 1, // Claude Sonnet is at index 1
	}
	m = m.loadStatus()
	return m
}

func (m WorktreeModel) Init() tea.Cmd {
	return nil
}

func (m WorktreeModel) loadStatus() WorktreeModel {
	m.isLoading = true
	m.loadingMsg = "Loading git status..."
	status, err := git.GetStatus(m.worktree.Path)
	m.isLoading = false
	if err != nil {
		m.err = err
		return m
	}
	m.status = status
	return m
}

func (m WorktreeModel) loadDiff() WorktreeModel {
	m.isLoading = true
	m.loadingMsg = "Loading diff..."
	diff, err := git.GetDiff(m.worktree.Path)
	m.isLoading = false
	if err != nil {
		m.err = err
		return m
	}
	m.diff = diff
	m.diffScroll = 0
	return m
}

func (m WorktreeModel) loadTicket() WorktreeModel {
	m.isLoading = true
	m.loadingMsg = "Loading ticket..."
	ticket, err := jira.LoadTicketCache(m.worktree.Ticket)
	if err != nil {
		m.err = err
		m.isLoading = false
		return m
	}
	if ticket == nil {
		return m.fetchTicket()
	}
	m.isLoading = false
	m.ticket = ticket
	m.ticketScroll = 0
	return m
}

func (m WorktreeModel) fetchTicket() WorktreeModel {
	m.isLoading = true
	m.loadingMsg = "Fetching from Jira..."
	ticket, err := jira.FetchTicket(m.worktree.Ticket)
	m.isLoading = false
	if err != nil {
		m.err = err
		return m
	}
	if err := jira.SaveTicketCache(ticket); err != nil {
		m.err = err
		return m
	}
	m.ticket = ticket
	m.ticketScroll = 0
	return m
}

func (m WorktreeModel) editTicket() WorktreeModel {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, m.getTicketCachePath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return m.loadTicket()
}

func (m WorktreeModel) appendToTicketNotes() WorktreeModel {
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
	return m.loadTicket()
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
	if m.promptModel || m.promptIter || m.selectingModel {
		return m.updateAgentPrompts(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "?" {
			m.showingHelp = !m.showingHelp
			return m, nil
		}
		if m.showingHelp && (msg.String() == "q" || msg.String() == "esc") {
			m.showingHelp = false
			return m, nil
		}
		if m.showingHelp {
			return m, nil
		}
		switch msg.String() {
		case "g":
			m.activeTab = TabChanges
			return m, nil
		case "a":
			if m.activeTab == TabTicket {
				m = m.appendToTicketNotes()
				return m, nil
			}
			m.activeTab = TabAgent
			return m, nil
		case "t":
			m.activeTab = TabTicket
			m = m.loadTicket()
			return m, nil
		case "p":
			m.activeTab = TabPlan
			m = m.loadPlan()
			return m, nil
		case "h", "left":
			m.activeTab = (m.activeTab - 1 + TabCount) % TabCount
			return m, nil
		case "l", "right":
			m.activeTab = (m.activeTab + 1) % TabCount
			return m, nil
		case "q", "esc":
			if m.activeTab == TabPlan && m.confirmPlan {
				m.confirmPlan = false
				return m, nil
			}
			if m.activeTab == TabPlan && m.confirmExecute {
				m.confirmExecute = false
				return m, nil
			}
			return m, func() tea.Msg { return WorktreeBackMsg{} }
		case "r":
			if m.activeTab == TabChanges {
				m = m.loadStatus()
			} else if m.activeTab == TabTicket {
				m = m.fetchTicket()
			} else if m.activeTab == TabAgent && !m.agentRunning {
				m.agentRunning = true
				m.agentDone = false
				return m, m.runAgentCmd("ralph")
			}
			return m, nil
		case "o":
			if m.activeTab == TabAgent && !m.agentRunning {
				m.agentRunning = true
				m.agentDone = false
				return m, m.runAgentCmd("opencode")
			}
			return m, nil
		case "m":
			if m.activeTab == TabAgent && !m.agentRunning {
				m.selectingModel = true
				// Find the current model's index
				for i, model := range AvailableModels {
					if model == m.agentModel {
						m.modelSelectorIdx = i
						break
					}
				}
			}
			return m, nil
		case "i":
			if m.activeTab == TabAgent && !m.agentRunning {
				m.promptIter = true
				m.agentIter = ""
			}
			return m, nil
		case "v":
			if m.activeTab == TabChanges {
				m.loadDiff()
				m.showingDiff = true
			} else if m.activeTab == TabAgent && m.agentDone {
				m.activeTab = TabChanges
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
					m = m.loadStatus()
				}
			} else if m.activeTab == TabAgent && m.agentRunning {
				exec.Command("tmux", "kill-session", "-t", m.worktree.Ticket).Run()
				m.agentRunning = false
				m.agentDone = true
			}
			return m, nil
		case "c":
			if m.activeTab == TabPlan {
				m.confirmPlan = true
				return m, nil
			}
			if m.activeTab == TabChanges && len(m.status) > 0 {
				m.commitMode = true
				m.commitMsg = fmt.Sprintf("[%s]: ", m.worktree.Ticket)
			}
			if m.activeTab == TabAgent && m.agentDone {
				m.activeTab = TabChanges
				m.commitMode = true
				m.commitMsg = fmt.Sprintf("[%s]: ", m.worktree.Ticket)
			}
			return m, nil
		case "e":
			if m.activeTab == TabTicket {
				m = m.editTicket()
			} else if m.activeTab == TabPlan && m.planExists {
				m = m.editPlan()
			}
			return m, nil
		case "j", "down":
			if m.activeTab == TabTicket {
				m.ticketScroll++
			} else if m.activeTab == TabPlan {
				m.planScroll++
			}
			return m, nil
		case "k", "up":
			if m.activeTab == TabTicket && m.ticketScroll > 0 {
				m.ticketScroll--
			} else if m.activeTab == TabPlan && m.planScroll > 0 {
				m.planScroll--
			}
			return m, nil
		case "x":
			if m.activeTab == TabPlan && m.planExists {
				m.confirmExecute = true
				return m, nil
			}
			return m, nil
		case "y":
			if m.activeTab == TabPlan && m.confirmPlan {
				m.isLoading = true
				m.loadingMsg = "Generating..."
				return m, m.generatePlanCmd()
			}
			if m.activeTab == TabPlan && m.confirmExecute {
				m.confirmExecute = false
				m.isLoading = true
				m.loadingMsg = "Starting OpenCode in tmux..."
				return m, m.executePlanInTmuxCmd()
			}
			return m, nil
		case "n":
			if m.activeTab == TabPlan && m.confirmPlan {
				m.confirmPlan = false
				return m, nil
			}
			if m.activeTab == TabPlan && m.confirmExecute {
				m.confirmExecute = false
				return m, nil
			}
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case planExecutionStartedMsg:
		m.isLoading = false
		m.activeTab = TabAgent
		return m, nil
	case planExecutionErrorMsg:
		m.isLoading = false
		m.err = msg.err
		return m, nil
	case planGeneratedMsg:
		m.isLoading = false
		m.confirmPlan = false
		m = m.loadPlan()
		return m, nil
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
				m = m.loadStatus()
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

func (m WorktreeModel) updateAgentPrompts(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.selectingModel {
				m.agentModel = AvailableModels[m.modelSelectorIdx]
				m.selectingModel = false
			} else if m.promptIter {
				m.promptIter = false
			}
			return m, nil
		case tea.KeyEscape:
			m.selectingModel = false
			m.promptIter = false
			return m, nil
		case tea.KeyUp:
			if m.selectingModel && m.modelSelectorIdx > 0 {
				m.modelSelectorIdx--
			}
			return m, nil
		case tea.KeyDown:
			if m.selectingModel && m.modelSelectorIdx < len(AvailableModels)-1 {
				m.modelSelectorIdx++
			}
			return m, nil
		case tea.KeyBackspace:
			if m.promptIter && len(m.agentIter) > 0 {
				m.agentIter = m.agentIter[:len(m.agentIter)-1]
			}
			return m, nil
		}
		if len(msg.String()) == 1 {
			if m.promptIter {
				m.agentIter += msg.String()
			}
		}
	}
	return m, nil
}

func (m *WorktreeModel) runAgentCmd(agentType string) tea.Cmd {
	return func() tea.Msg {
		sessionName := m.worktree.Ticket
		exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", m.worktree.Path).Run()

		var cmdStr string
		if agentType == "opencode" {
			cmdStr = fmt.Sprintf("opencode run --model %s 'use ralph-implementer skill'\n", m.agentModel)
		} else {
			cmdStr = fmt.Sprintf("ralph --model %s --iterations %s\n", m.agentModel, m.agentIter)
		}

		exec.Command("tmux", "send-keys", "-t", sessionName, cmdStr).Run()
		return nil
	}
}

func (m WorktreeModel) renderAgentTab() string {
	if m.selectingModel {
		var lines []string
		lines = append(lines, promptStyle.Render("  Select Model (↑/↓ to navigate, Enter to select):"))
		lines = append(lines, "")
		for i, model := range AvailableModels {
			if i == m.modelSelectorIdx {
				lines = append(lines, tabActiveStyle.Render(fmt.Sprintf("  > %s", model)))
			} else {
				lines = append(lines, fmt.Sprintf("    %s", model))
			}
		}
		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}
	if m.promptIter {
		return promptStyle.Render(fmt.Sprintf("  Iterations: %s", m.agentIter))
	}

	var lines []string
	if m.agentRunning {
		lines = append(lines, infoTextStyle.Render(fmt.Sprintf("  Running in tmux session: %s", m.worktree.Ticket)))
		lines = append(lines, infoTextStyle.Render(fmt.Sprintf("  tmux attach -t %s", m.worktree.Ticket)))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("  [s] Stop/kill tmux session"))
	} else if m.agentDone {
		lines = append(lines, infoTextStyle.Render("  Agent finished or stopped."))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("  [v] View Changes  [c] Commit Changes"))
	} else {
		lines = append(lines, infoTextStyle.Render("  Agent Ready"))
		lines = append(lines, infoTextStyle.Render(fmt.Sprintf("  Current Model: %s", m.agentModel)))
		lines = append(lines, infoTextStyle.Render(fmt.Sprintf("  Iterations: %s", m.agentIter)))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("  [o] Run OpenCode  [r] Run Ralph  [m] Set Model  [i] Set Iterations"))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m WorktreeModel) View() string {
	if m.showingHelp {
		return m.renderHelp()
	}
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
		return m.renderAgentTab()
	case TabTicket:
		return m.renderTicketTab()
	case TabPlan:
		return m.renderPlanTab()
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

func (m WorktreeModel) getPlanPath() string {
	homeDir, _ := os.UserHomeDir()
	return fmt.Sprintf("%s/.scratch/tickets/%s_plan.md", homeDir, m.worktree.Ticket)
}

func (m WorktreeModel) loadPlan() WorktreeModel {
	path := m.getPlanPath()
	content, err := os.ReadFile(path)
	if err != nil {
		m.planExists = false
		m.planContent = "No plan file found"
	} else {
		m.planExists = true
		m.planContent = string(content)
	}
	m.planScroll = 0
	return m
}

func (m WorktreeModel) editPlan() WorktreeModel {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	cmd := exec.Command(editor, m.getPlanPath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return m.loadPlan()
}

type planGeneratedMsg struct{}

func (m *WorktreeModel) generatePlanCmd() tea.Cmd {
	return func() tea.Msg {
		path := m.getPlanPath()
		prompt := fmt.Sprintf("Read ~/.scratch/tickets/%s.md and write a step-by-step implementation plan to %s. DO NOT use the edit or write tools to write anything besides the plan file. DO NOT do any other tasks. Provide extreme detail. Only output the markdown for the plan.", m.worktree.Ticket, path)
		cmd := exec.Command("opencode", "run", prompt)
		cmd.Run()
		return planGeneratedMsg{}
	}
}

type planExecutionStartedMsg struct{}
type planExecutionErrorMsg struct{ err error }

func (m *WorktreeModel) executePlanInTmuxCmd() tea.Cmd {
	return func() tea.Msg {
		sessionName := m.worktree.Ticket

		// Check if tmux session already exists
		checkCmd := exec.Command("tmux", "has-session", "-t", sessionName)
		if err := checkCmd.Run(); err != nil {
			// Session doesn't exist, create it in detached mode
			createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", m.worktree.Path)
			if err := createCmd.Run(); err != nil {
				return planExecutionErrorMsg{err}
			}
		}

		// Send command to tmux session
		cmdStr := "opencode run 'use ralph-implementer skill'\n"
		sendCmd := exec.Command("tmux", "send-keys", "-t", sessionName, cmdStr)
		if err := sendCmd.Run(); err != nil {
			return planExecutionErrorMsg{err}
		}

		return planExecutionStartedMsg{}
	}
}

func (m WorktreeModel) renderPlanTab() string {
	if m.confirmPlan {
		return promptStyle.Render("  Generate new plan with AI? (y/n)")
	}
	if m.confirmExecute {
		return promptStyle.Render("  Execute plan with OpenCode in tmux? (y/n)")
	}

	var lines []string

	if !m.planExists {
		lines = append(lines, infoTextStyle.Render("  No plan file found"))
	} else {
		lines = append(lines, ticketLabelStyle.Render("  Implementation Plan:"))
		lines = append(lines, ticketLabelStyle.Render("  ─────────────────────────────────────────────────"))

		planLines := strings.Split(m.planContent, "\n")
		start := m.planScroll
		end := start + 10
		if end > len(planLines) {
			end = len(planLines)
		}

		if start > 0 {
			lines = append(lines, infoTextStyle.Render("  ... (scroll up with k)"))
		}
		for i := start; i < end; i++ {
			lines = append(lines, ticketValueStyle.Render("  "+planLines[i]))
		}
		if end < len(planLines) {
			lines = append(lines, infoTextStyle.Render("  ... (scroll down with j)"))
		}
		lines = append(lines, ticketLabelStyle.Render("  ─────────────────────────────────────────────────"))
	}
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  [c] Generate Plan  [e] Edit Plan  [x] Execute Plan"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m WorktreeModel) renderHelp() string {
	var lines []string
	lines = append(lines, headerStyle.Render("Keyboard Shortcuts - Worktree Details"))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Global:"))
	lines = append(lines, "  ?:        Toggle help")
	lines = append(lines, "  q/Esc:    Go back / cancel")
	lines = append(lines, "  h/l:      Previous / next tab")
	lines = append(lines, "  g/a/t/p:  Jump to tab (Changes/Agent/Ticket/Plan)")
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Changes Tab:"))
	lines = append(lines, "  v:        View diff")
	lines = append(lines, "  s:        Stage all changes")
	lines = append(lines, "  c:        Commit staged changes")
	lines = append(lines, "  r:        Refresh status")
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Agent Tab:"))
	lines = append(lines, "  r:        Run Ralph")
	lines = append(lines, "  o:        Run OpenCode")
	lines = append(lines, "  s:        Stop agent")
	lines = append(lines, "  m:        Change model")
	lines = append(lines, "  i:        Change iterations")
	lines = append(lines, "  v:        View changes (when done)")
	lines = append(lines, "  c:        Commit changes (when done)")
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Ticket Tab:"))
	lines = append(lines, "  r:        Refetch ticket")
	lines = append(lines, "  a:        Append notes")
	lines = append(lines, "  e:        Edit raw file")
	lines = append(lines, "  j/k:      Scroll")
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  Plan Tab:"))
	lines = append(lines, "  c:        Generate plan")
	lines = append(lines, "  e:        Edit plan")
	lines = append(lines, "  x:        Execute plan")
	lines = append(lines, "  j/k:      Scroll")
	lines = append(lines, "")
	lines = append(lines, "Press ? or Esc to close help")
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
