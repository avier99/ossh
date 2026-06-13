package ssh

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestConnect_EmptyHostAlias(t *testing.T) {
	err := Connect("")
	if !errors.Is(err, ErrEmptyHostAlias) {
		t.Errorf("expected ErrEmptyHostAlias, got %v", err)
	}
}

func TestConnect_SSHNotOnPath(t *testing.T) {
	// Save the original PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Set PATH to empty so ssh won't be found
	os.Setenv("PATH", "")

	// Verify ssh is not found
	_, err := exec.LookPath("ssh")
	if err == nil {
		t.Skip("ssh still found even with empty PATH, skipping test")
	}

	err = Connect("test-host")
	if !errors.Is(err, ErrSSHNotFound) {
		t.Errorf("expected ErrSSHNotFound, got %v", err)
	}
}

// Note: syscall.Exec itself cannot be unit tested because it replaces the current process.
// On success, Exec never returns — the process becomes ssh.
// Integration tests would need to spawn a separate process to verify the handoff,
// but that is out of scope for unit tests.
