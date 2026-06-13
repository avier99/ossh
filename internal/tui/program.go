package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinburke/ssh_config"
	"github.com/synoxisllc/ossh/internal/config"
)

// Start launches the TUI with the given initial screen, SSH config, and audit findings.
// It blocks until the user quits or initiates an SSH connection.
// Returns any error from the SSH connection attempt.
func Start(initialScreen Screen, cfg *ssh_config.Config, findings []config.Finding) error {
	// Create root model
	model := Model{
		screen:   initialScreen,
		cfg:      cfg,
		findings: findings,
	}

	// Create program with alt screen (keeps terminal clean on exit)
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Run blocks until quit
	finalModel, err := program.Run()
	if err != nil {
		return err
	}

	// If we quit due to a connect attempt, return any connect error
	if m, ok := finalModel.(Model); ok {
		return m.connectErr
	}

	return nil
}
