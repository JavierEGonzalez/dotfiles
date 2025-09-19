package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	tickets  []string         // list of ticket IDs
	cursor   int              // which ticket is selected
	selected map[int]struct{} // which tickets are selected
}

func initialModel() model {
	return model{
		tickets: []string{},

		// A map which indicates which tickets are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `tickets` slice, above.
		selected: make(map[int]struct{}),
	}
}
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.tickets)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}
func (m model) View() string {
	s := "Branches with tickets:\n"

	cmd := exec.Command("git", "branch", "-r")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	re := regexp.MustCompile(`[a-zA-Z]+-[0-9]+`)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		matches := re.FindAllString(line, -1)
		for _, match := range matches {
			m.tickets = append(m.tickets, match)
		}
	}
	// remove duplicates
	ticketSet := make(map[string]struct{})
	for _, ticket := range m.tickets {
		ticketSet[ticket] = struct{}{}
	}

	// Iterate over our tickets
	for i, choice := range m.tickets {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
