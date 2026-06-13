package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/synoxisllc/ossh/internal/config"
	"github.com/synoxisllc/ossh/internal/tui"
	"github.com/synoxisllc/ossh/internal/tui/screens"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get SSH directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	sshDir := filepath.Join(home, ".ssh")

	// First-run setup (creates ~/.ssh/ if needed, fixes perms)
	if err := config.Setup(sshDir); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Acquire lock
	release, err := config.AcquireLock(sshDir)
	if err != nil {
		if errors.Is(err, config.ErrLockHeld) {
			fmt.Fprintln(os.Stderr, "Another ossh instance is running. Please wait or cancel that instance.")
			return nil
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer release()

	// Load SSH config
	configPath := filepath.Join(sshDir, "config")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load SSH config: %w", err)
	}

	// Audit permissions
	findings, err := config.Audit(sshDir)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	// Create initial home screen
	homeScreen := screens.NewHome(cfg, findings)

	// Start TUI
	return tui.Start(homeScreen, cfg, findings)
}
