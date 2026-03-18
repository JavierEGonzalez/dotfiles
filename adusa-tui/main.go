package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/javiergonzalez/adusa-tui/internal/git"
	"github.com/javiergonzalez/adusa-tui/internal/ui/screens"
)

const (
	ViewWorktrees = "worktrees"
	ViewCreate    = "create"
	ViewDelete    = "delete"
	ViewWorktree  = "worktree"

	// Terminal size threshold (in character cells) for applying 80% scaling
	thresholdCols = 150
	thresholdRows = 45
)

var (
	containerBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#45475a"))
	scrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6c7086"))
)

type model struct {
	worktreesScreen      screens.WorktreesModel
	createWorktreeScreen screens.CreateWorktreeModel
	deleteWorktreeScreen screens.DeleteWorktreeModel
	worktreeScreen       screens.WorktreeModel
	viewport             viewport.Model
	view                 string
	width                int
	height               int
	initTicket           string
	initProcessed        bool
	ready                bool
}

// containerSize returns the content area width and height based on terminal dimensions.
// If the terminal is >= 150 cols and >= 45 rows, use 80% of each dimension.
// Otherwise, use the full terminal size.
func (m model) containerSize() (int, int) {
	w, h := m.width, m.height
	if w >= thresholdCols && h >= thresholdRows {
		w = w * 80 / 100
		h = h * 80 / 100
	}
	return w, h
}

// innerSize returns the usable content area inside the bordered container.
func (m model) innerSize() (int, int) {
	cw, ch := m.containerSize()
	// Account for border (1 char each side)
	w := cw - 2
	h := ch - 2
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m *model) processInit() tea.Cmd {
	if m.initTicket != "" && !m.initProcessed {
		m.initProcessed = true
		wt := m.worktreesScreen.FindByPath(m.initTicket)
		if wt == nil {
			wt = m.worktreesScreen.FindByTicket(m.initTicket)
		}
		if wt != nil {
			m.view = ViewWorktree
			m.worktreeScreen = screens.NewWorktreeModel(*wt)
			return m.worktreeScreen.Init()
		}
	}
	return nil
}

func (m *model) initViewport() {
	w, h := m.innerSize()
	m.viewport = viewport.New(w, h)
	m.viewport.Style = lipgloss.NewStyle()
	// Disable all keyboard bindings — screens handle their own navigation.
	// The viewport is only used for mouse-wheel scrolling and as a render container.
	m.viewport.KeyMap = viewport.KeyMap{
		PageDown:     key.NewBinding(key.WithKeys("pgdown")),
		PageUp:       key.NewBinding(key.WithKeys("pgup")),
		HalfPageUp:   key.NewBinding(key.WithDisabled()),
		HalfPageDown: key.NewBinding(key.WithDisabled()),
		Up:           key.NewBinding(key.WithDisabled()),
		Down:         key.NewBinding(key.WithDisabled()),
		Left:         key.NewBinding(key.WithDisabled()),
		Right:        key.NewBinding(key.WithDisabled()),
	}
	m.ready = true
}

func (m *model) resizeViewport() {
	w, h := m.innerSize()
	m.viewport.Width = w
	m.viewport.Height = h
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.initViewport()
		} else {
			m.resizeViewport()
		}
		innerW, _ := m.innerSize()
		m.worktreesScreen.SetWidth(innerW)
		cmd := m.processInit()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	// Forward mouse events to the viewport for scroll-wheel support
	switch msg.(type) {
	case tea.MouseMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Delegate to the active screen
	var cmd tea.Cmd
	switch m.view {
	case ViewCreate:
		m, cmd = m.delegateCreateWorktree(msg)
	case ViewDelete:
		m, cmd = m.delegateDeleteWorktree(msg)
	case ViewWorktree:
		m, cmd = m.delegateWorktree(msg)
	default:
		m, cmd = m.delegateWorktrees(msg)
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) delegateWorktrees(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.WorktreeSelectedMsg:
		m.view = ViewWorktree
		m.worktreeScreen = screens.NewWorktreeModel(msg.Worktree)
		return m, m.worktreeScreen.Init()
	case screens.CreateWorktreeMsg:
		m.view = ViewCreate
		m.createWorktreeScreen = screens.NewCreateWorktreeModel()
		return m, nil
	case screens.DeleteWorktreeMsg:
		m.view = ViewDelete
		m.deleteWorktreeScreen = screens.NewDeleteWorktreeModel(msg.Worktree)
		return m, nil
	}

	newScreen, cmd := m.worktreesScreen.Update(msg)
	m.worktreesScreen = newScreen.(screens.WorktreesModel)
	return m, cmd
}

func (m model) delegateCreateWorktree(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.CreateWorktreeDoneMsg:
		m.view = ViewWorktrees
		wt, _ := m.worktreesScreen.Refresh()
		m.worktreesScreen = wt
		return m, nil
	default:
		newScreen, cmd := m.createWorktreeScreen.Update(msg)
		m.createWorktreeScreen = newScreen.(screens.CreateWorktreeModel)
		return m, cmd
	}
}

func (m model) delegateDeleteWorktree(msg tea.Msg) (model, tea.Cmd) {
	switch msg.(type) {
	case screens.DeleteWorktreeDoneMsg:
		m.view = ViewWorktrees
		wt, _ := m.worktreesScreen.Refresh()
		m.worktreesScreen = wt
		return m, nil
	case screens.DeleteWorktreeCancelledMsg:
		m.view = ViewWorktrees
		return m, nil
	default:
		newScreen, cmd := m.deleteWorktreeScreen.Update(msg)
		m.deleteWorktreeScreen = newScreen.(screens.DeleteWorktreeModel)
		return m, cmd
	}
}

func (m model) delegateWorktree(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.WorktreeBackMsg:
		m.view = ViewWorktrees
		return m, nil
	default:
		newScreen, cmd := m.worktreeScreen.Update(msg)
		m.worktreeScreen = newScreen.(screens.WorktreeModel)
		return m, cmd
	}
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 || !m.ready {
		return ""
	}

	var content string
	switch m.view {
	case ViewCreate:
		content = m.createWorktreeScreen.View()
	case ViewDelete:
		content = m.deleteWorktreeScreen.View()
	case ViewWorktree:
		content = m.worktreeScreen.View()
	default:
		content = m.worktreesScreen.View()
	}

	innerW, innerH := m.innerSize()

	// Update viewport content for rendering
	m.viewport.Width = innerW
	m.viewport.Height = innerH
	m.viewport.SetContent(content)

	// Count content lines to decide if we need a scroll indicator
	contentLines := strings.Count(content, "\n") + 1
	scrollable := contentLines > innerH

	// Build the bordered container
	bordered := containerBorderStyle.
		Width(innerW).
		Height(innerH).
		Render(m.viewport.View())

	// Add scroll indicator below the border if content overflows
	if scrollable {
		cw, _ := m.containerSize()
		pct := m.viewport.ScrollPercent()
		indicator := scrollIndicatorStyle.Render(fmt.Sprintf(" %3.0f%% ", pct*100))
		bordered = bordered + "\n" + lipgloss.PlaceHorizontal(cw, lipgloss.Center, indicator)
	}

	// Center the container in the full terminal
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

func main() {
	initialTicket := ""
	if len(os.Args) > 1 {
		initialTicket = os.Args[1]
	}

	worktreesScreen := screens.NewWorktreesModel()

	if initialTicket == "" {
		if wt, err := git.GetCurrentWorktree(); err == nil && wt != nil {
			initialTicket = wt.Path
		}
	}

	if _, err := tea.NewProgram(model{
		worktreesScreen: worktreesScreen,
		initTicket:      initialTicket,
	}, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
