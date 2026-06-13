package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/kevinburke/ssh_config"
	"github.com/synoxisllc/ossh/internal/config"
	"github.com/synoxisllc/ossh/internal/ssh"
)

// Screen is the interface that all TUI screens must implement.
// Screens are Elm-architecture components managed by the root Model.
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

// Model is the root TUI model.
// It holds shared state and delegates to the current active screen.
type Model struct {
	screen   Screen
	cfg      *ssh_config.Config
	findings []config.Finding
	width    int
	height   int

	// connectErr stores any error from ssh.Connect so it can be returned
	// after the program quits
	connectErr error
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.screen != nil {
		return m.screen.Init()
	}
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to current screen
		if m.screen != nil {
			var cmd tea.Cmd
			m.screen, cmd = m.screen.Update(msg)
			return m, cmd
		}
		return m, nil

	case SwitchScreenMsg:
		m.screen = msg.Next
		return m, m.screen.Init()

	case ConnectMsg:
		// Attempt to connect. Since syscall.Exec replaces the process,
		// if this returns, it means the exec failed.
		m.connectErr = ssh.Connect(msg.HostAlias)
		return m, tea.Quit

	default:
		// Delegate to current screen
		if m.screen != nil {
			var cmd tea.Cmd
			m.screen, cmd = m.screen.Update(msg)
			return m, cmd
		}
		return m, nil
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.screen != nil {
		return m.screen.View()
	}
	return ""
}
