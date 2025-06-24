package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)
)

type state int

const (
	stateInput state = iota
	stateBranchType
	stateDescription
	stateBranchSelect
	stateComplete
)

type branchItem struct {
	name string
}

func (i branchItem) FilterValue() string { return i.name }
func (i branchItem) Title() string       { return i.name }
func (i branchItem) Description() string { return "" }

type model struct {
	state       state
	ticketInput textinput.Model
	descInput   textinput.Model
	branchList  list.Model
	ticket      string
	branchType  string
	description string
	branches    []string
	message     string
	err         error
	quitting    bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter ticket number (e.g., 123)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	di := textinput.New()
	di.Placeholder = "Enter branch description (optional)"
	di.CharLimit = 100
	di.Width = 50

	return model{
		state:       stateInput,
		ticketInput: ti,
		descInput:   di,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				if m.ticketInput.Value() == "" {
					m.err = fmt.Errorf("ticket number cannot be empty")
					return m, nil
				}
				m.ticket = "CXPVSP-" + m.ticketInput.Value()
				m.err = nil
				
				// Process the ticket
				if err := m.processTicket(); err != nil {
					m.err = err
					return m, nil
				}
				
				// Check if we need to create or select branches
				branches, err := m.getBranches()
				if err != nil {
					m.err = err
					return m, nil
				}
				
				if len(branches) == 0 {
					m.state = stateBranchType
					m.message = "No branches found for this ticket"
				} else {
					m.branches = branches
					m.setupBranchList()
					m.state = stateBranchSelect
					m.message = "Select a branch to checkout:"
				}
				return m, nil
			}

		case stateBranchType:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "f", "F":
				m.branchType = "feature"
				m.state = stateDescription
				m.descInput.Focus()
				return m, textinput.Blink
			case "b", "B":
				m.branchType = "bugfix"
				m.state = stateDescription
				m.descInput.Focus()
				return m, textinput.Blink
			case "h", "H":
				m.branchType = "hotfix"
				m.state = stateDescription
				m.descInput.Focus()
				return m, textinput.Blink
			case "n", "N":
				m.state = stateComplete
				m.message = "Skipped branch creation"
				return m, nil
			}

		case stateDescription:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				m.description = m.descInput.Value()
				if err := m.createBranch(); err != nil {
					m.err = err
					return m, nil
				}
				m.state = stateComplete
				return m, nil
			}

		case stateBranchSelect:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				selected := m.branchList.SelectedItem()
				if selected != nil {
					branchName := selected.(branchItem).name
					if branchName != "Skip checkout" {
						if err := m.checkoutBranch(branchName); err != nil {
							m.err = err
							return m, nil
						}
						m.message = fmt.Sprintf("Checked out branch: %s", branchName)
					} else {
						m.message = "Skipped branch checkout"
					}
				}
				m.state = stateComplete
				return m, nil
			}

		case stateComplete:
			switch msg.String() {
			case "ctrl+c", "q", "enter":
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case stateInput:
		m.ticketInput, cmd = m.ticketInput.Update(msg)
	case stateDescription:
		m.descInput, cmd = m.descInput.Update(msg)
	case stateBranchSelect:
		m.branchList, cmd = m.branchList.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s strings.Builder
	s.WriteString(titleStyle.Render("🎫 Ticket Manager") + "\n\n")

	switch m.state {
	case stateInput:
		s.WriteString("Enter ticket number:\n")
		s.WriteString(m.ticketInput.View() + "\n\n")
		s.WriteString(helpStyle.Render("Press Enter to continue, q to quit"))

	case stateBranchType:
		s.WriteString(fmt.Sprintf("Ticket: %s\n", statusStyle.Render(m.ticket)))
		s.WriteString(m.message + "\n\n")
		s.WriteString("Create a new branch?\n")
		s.WriteString("  [f] Feature branch\n")
		s.WriteString("  [b] Bugfix branch\n")
		s.WriteString("  [h] Hotfix branch\n")
		s.WriteString("  [n] No, skip\n\n")
		s.WriteString(helpStyle.Render("Select branch type or press n to skip, q to quit"))

	case stateDescription:
		s.WriteString(fmt.Sprintf("Ticket: %s\n", statusStyle.Render(m.ticket)))
		s.WriteString(fmt.Sprintf("Type: %s\n\n", statusStyle.Render(m.branchType)))
		s.WriteString("Branch description (optional):\n")
		s.WriteString(m.descInput.View() + "\n\n")
		s.WriteString(helpStyle.Render("Press Enter to create branch, q to quit"))

	case stateBranchSelect:
		s.WriteString(fmt.Sprintf("Ticket: %s\n", statusStyle.Render(m.ticket)))
		s.WriteString(m.message + "\n\n")
		s.WriteString(m.branchList.View())
		s.WriteString("\n" + helpStyle.Render("Use ↑/↓ to navigate, Enter to select, q to quit"))

	case stateComplete:
		s.WriteString(fmt.Sprintf("Ticket: %s\n", statusStyle.Render(m.ticket)))
		if m.message != "" {
			s.WriteString(statusStyle.Render("✓ " + m.message) + "\n")
		}
		s.WriteString("\n" + helpStyle.Render("Press Enter or q to exit"))
	}

	if m.err != nil {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.err.Error()))
	}

	return s.String()
}

func (m *model) processTicket() error {
	// Set tmux environment variable
	cmd := exec.Command("tmux", "setenv", "ticket", m.ticket)
	if err := cmd.Run(); err != nil {
		// Don't fail if tmux isn't running
		fmt.Printf("Warning: Could not set tmux environment: %v\n", err)
	}

	// Add to tickets file
	return m.addToTicketsFile()
}

func (m *model) addToTicketsFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	ticketsFile := filepath.Join(homeDir, ".scratch", ".currentTickets.txt")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(ticketsFile), 0755); err != nil {
		return err
	}

	// Check if ticket already exists
	if _, err := os.Stat(ticketsFile); err == nil {
		file, err := os.Open(ticketsFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) == m.ticket {
				m.message = "Ticket already in file"
				return nil
			}
		}
	}

	// Read existing content
	var lines []string
	if _, err := os.Stat(ticketsFile); err == nil {
		file, err := os.Open(ticketsFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	}

	// Write new content with ticket at the top
	file, err := os.Create(ticketsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	fmt.Fprintln(writer, m.ticket)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()

	m.message = fmt.Sprintf("Added %s to tickets file", m.ticket)
	return nil
}

func (m *model) getBranches() ([]string, error) {
	cmd := exec.Command("git", "branch")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git branches: %v", err)
	}

	var branches []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ")
		if line != "" && strings.Contains(line, m.ticket) {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

func (m *model) setupBranchList() {
	items := make([]list.Item, 0, len(m.branches)+1)
	for _, branch := range m.branches {
		items = append(items, branchItem{name: branch})
	}
	items = append(items, branchItem{name: "Skip checkout"})

	m.branchList = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.branchList.Title = "Available Branches"
	m.branchList.SetHeight(10)
	m.branchList.SetWidth(60)
}

func (m *model) createBranch() error {
	branchName := fmt.Sprintf("%s/%s", m.branchType, m.ticket)
	if m.description != "" {
		desc := strings.ReplaceAll(m.description, " ", "-")
		branchName += "-" + desc
	}

	cmd := exec.Command("git", "checkout", "-b", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	m.message = fmt.Sprintf("Created and checked out branch: %s", branchName)
	return nil
}

func (m *model) checkoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch: %v", err)
	}
	return nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
