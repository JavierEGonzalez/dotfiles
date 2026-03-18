package screens

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/config"
	"github.com/javiergonzalez/adusa-tui/internal/git"
	"github.com/javiergonzalez/adusa-tui/internal/jira"
)

var (
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89b4fa"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f38ba8"))
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1"))
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086"))
)

type CreateWorktreeModel struct {
	step               int
	repoInput          textinput.Model
	ticketInput        textinput.Model
	branchInput        textinput.Model
	descInput          textinput.Model
	repo               string
	repoPath           string
	worktreeDir        string
	ticket             string
	branchType         string
	description        string
	descriptionDefault string
	err                error
	success            bool
}

func NewCreateWorktreeModel() CreateWorktreeModel {
	repos := config.AppConfig.Repos
	defaultRepo := ""
	defaultPath := ""
	defaultWorktreeDir := ""
	if len(repos) > 0 {
		defaultRepo = repos[0].Name
		defaultPath = repos[0].Path
		defaultWorktreeDir = repos[0].GetWorktreeDir()
	}

	repo := textinput.New()
	repo.Placeholder = defaultRepo + " (or enter repo name)"
	repo.Prompt = "Repo: "

	ticket := textinput.New()
	ticket.Placeholder = "CXPVSP-123"
	ticket.Prompt = "Ticket: "

	branch := textinput.New()
	branch.Placeholder = "f (feature), b (bugfix), h (hotfix)"
	branch.Prompt = "Branch: "

	desc := textinput.New()
	desc.Placeholder = "optional description"
	desc.Prompt = "Description: "

	return CreateWorktreeModel{
		step:        0,
		repoInput:   repo,
		ticketInput: ticket,
		branchInput: branch,
		descInput:   desc,
		repo:        defaultRepo,
		repoPath:    defaultPath,
		worktreeDir: defaultWorktreeDir,
	}
}

func (m CreateWorktreeModel) Init() tea.Cmd {
	return nil
}

func (m CreateWorktreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.success || m.err != nil {
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.Type == tea.KeyEnter || key.Type == tea.KeyEsc {
				return m, func() tea.Msg { return CreateWorktreeDoneMsg{} }
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return m, func() tea.Msg { return CreateWorktreeDoneMsg{} }
		case tea.KeyEnter:
			return m.handleEnter()
		}
	}

	switch m.step {
	case 0:
		newInput, cmd := m.repoInput.Update(msg)
		m.repoInput = newInput
		return m, cmd
	case 1:
		newInput, cmd := m.ticketInput.Update(msg)
		m.ticketInput = newInput
		return m, cmd
	case 2:
		newInput, cmd := m.branchInput.Update(msg)
		m.branchInput = newInput
		return m, cmd
	case 3:
		newInput, cmd := m.descInput.Update(msg)
		m.descInput = newInput
		return m, cmd
	}

	return m, nil
}

func (m CreateWorktreeModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		repoInput := strings.TrimSpace(m.repoInput.Value())
		if repoInput != "" {
			found := false
			for _, r := range config.AppConfig.Repos {
				if r.Name == repoInput {
					m.repo = r.Name
					m.repoPath = r.Path
					m.worktreeDir = r.GetWorktreeDir()
					found = true
					break
				}
			}
			if !found {
				m.err = fmt.Errorf("repo not found: %s", repoInput)
				return m, nil
			}
		}
		m.step = 1
		m.ticketInput.Focus()
		return m, nil
	case 1:
		ticket := strings.TrimSpace(m.ticketInput.Value())
		if ticket == "" {
			m.err = fmt.Errorf("ticket number is required")
			return m, nil
		}
		validTicket := regexp.MustCompile(`^[A-Z]+-\d+$`)
		if !validTicket.MatchString(ticket) {
			m.err = fmt.Errorf("invalid ticket format (use: CXPVSP-123)")
			return m, nil
		}
		m.ticket = ticket
		m.descriptionDefault = ""
		if cached, err := jira.LoadTicketCache(ticket); err == nil && cached != nil {
			m.descriptionDefault = strings.TrimSpace(cached.Summary)
		} else if fetched, err := jira.FetchTicket(ticket); err == nil && fetched != nil {
			m.descriptionDefault = strings.TrimSpace(fetched.Summary)
			_ = jira.SaveTicketCache(fetched)
		}
		if m.descriptionDefault != "" {
			m.descInput.Placeholder = m.descriptionDefault
		}
		m.step = 2
		m.branchInput.Focus()
		return m, nil
	case 2:
		branchType := strings.TrimSpace(m.branchInput.Value())
		if branchType == "" {
			m.err = fmt.Errorf("branch type is required (f/b/h)")
			return m, nil
		}
		if branchType != "f" && branchType != "b" && branchType != "h" {
			m.err = fmt.Errorf("invalid branch type (use: f=feature, b=bugfix, h=hotfix)")
			return m, nil
		}
		m.branchType = branchType
		m.step = 3
		m.descInput.Focus()
		return m, nil
	case 3:
		m.description = strings.TrimSpace(m.descInput.Value())
		if m.description == "" {
			m.description = strings.TrimSpace(m.descriptionDefault)
		}
		m.err = m.createWorktree()
		if m.err == nil {
			m.success = true
		}
		return m, nil
	}
	return m, nil
}

func (m CreateWorktreeModel) createWorktree() error {
	_, err := git.CreateWorktree(m.ticket, m.branchType, m.description, m.repoPath)
	return err
}

func (m CreateWorktreeModel) View() string {
	if m.success {
		dirName := m.ticket
		previewPath := fmt.Sprintf("%s/%s", m.worktreeDir, dirName)
		return fmt.Sprintf("\n  %s\n\n  Worktree created at: %s\n\n  Press Enter to continue...",
			successStyle.Render("✓ Worktree created successfully"),
			previewPath)
	}

	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n  Press Enter to continue...",
			errorStyle.Render("✗ Error"),
			errorStyle.Render(m.err.Error()))
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Create New Worktree"))
	lines = append(lines, "")

	switch m.step {
	case 0:
		repoNames := make([]string, len(config.AppConfig.Repos))
		for i, r := range config.AppConfig.Repos {
			repoNames[i] = r.Name
		}
		lines = append(lines, labelStyle.Render(fmt.Sprintf("Select repo (%s)", strings.Join(repoNames, ", "))))
		lines = append(lines, inputStyle.Render(m.repoInput.View()))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Press Enter to continue, Esc to cancel"))
	case 1:
		lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Repo"), m.repo))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Enter ticket number (e.g., CXPVSP-123)"))
		lines = append(lines, inputStyle.Render(m.ticketInput.View()))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Press Enter to continue, Esc to cancel"))
	case 2:
		lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Repo"), m.repo))
		lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Ticket"), m.ticket))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Select branch type"))
		lines = append(lines, labelStyle.Render("  f = feature, b = bugfix, h = hotfix"))
		lines = append(lines, inputStyle.Render(m.branchInput.View()))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Press Enter to continue, Esc to cancel"))
	case 3:
		dirName := m.ticket
		previewPath := fmt.Sprintf("%s/%s", m.worktreeDir, dirName)
		branchTypeName := map[string]string{"f": "feature", "b": "bugfix", "h": "hotfix"}[m.branchType]
		desc := strings.TrimSpace(m.descInput.Value())
		if desc == "" {
			desc = strings.TrimSpace(m.descriptionDefault)
		}
		slug := git.SlugifyDescription(desc)
		branchName := ""
		if slug == "" {
			branchName = fmt.Sprintf("%s/%s", branchTypeName, dirName)
		} else {
			branchName = fmt.Sprintf("%s/%s-%s", branchTypeName, dirName, slug)
		}

		lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Repo"), m.repo))
		lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Ticket"), m.ticket))
		lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Branch Type"), branchTypeName))
		if desc != "" {
			lines = append(lines, fmt.Sprintf("  %s: %s", labelStyle.Render("Description"), desc))
		}
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render(fmt.Sprintf("Preview: Worktree: %s", previewPath)))
		lines = append(lines, labelStyle.Render(fmt.Sprintf("Branch: %s", branchName)))
		lines = append(lines, "")
		lines = append(lines, inputStyle.Render(m.descInput.View()))
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Press Enter to create, Esc to cancel"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type CreateWorktreeDoneMsg struct{}
