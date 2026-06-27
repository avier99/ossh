package keys

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarget_WithUser(t *testing.T) {
	opts := CopyKeyOpts{User: "gana", HostAlias: "homelab"}
	got := Target(opts)
	if got != "gana@homelab" {
		t.Errorf("Target() = %q, want gana@homelab", got)
	}
}

func TestTarget_NoUser(t *testing.T) {
	opts := CopyKeyOpts{HostAlias: "homelab"}
	got := Target(opts)
	if got != "homelab" {
		t.Errorf("Target() = %q, want homelab", got)
	}
}

func TestManualFallback_ContainsPublicKey(t *testing.T) {
	opts := CopyKeyOpts{
		Key:       KeyEntry{DisplayPath: "~/.ssh/mykey"},
		HostAlias: "homelab",
	}
	got := ManualFallback(opts)
	if !strings.Contains(got, "~/.ssh/mykey.pub") {
		t.Errorf("ManualFallback() = %q, want to contain ~/.ssh/mykey.pub", got)
	}
}

func TestManualFallback_ContainsTarget(t *testing.T) {
	opts := CopyKeyOpts{
		Key:       KeyEntry{DisplayPath: "~/.ssh/mykey"},
		HostAlias: "homelab",
		User:      "gana",
	}
	got := ManualFallback(opts)
	if !strings.Contains(got, "gana@homelab") {
		t.Errorf("ManualFallback() = %q, want to contain gana@homelab", got)
	}
}

func TestManualFallback_ContainsAuthorizedKeys(t *testing.T) {
	opts := CopyKeyOpts{
		Key:       KeyEntry{DisplayPath: "~/.ssh/mykey"},
		HostAlias: "homelab",
	}
	got := ManualFallback(opts)
	if !strings.Contains(got, "authorized_keys") {
		t.Errorf("ManualFallback() = %q, want to contain authorized_keys", got)
	}
}

func TestCopyCommand_NotFound(t *testing.T) {
	t.Setenv("PATH", "")
	opts := CopyKeyOpts{
		SSHDir:    "/tmp",
		Key:       KeyEntry{RelPath: "mykey"},
		HostAlias: "homelab",
	}
	_, err := CopyCommand(opts)
	if !errors.Is(err, ErrSSHCopyIDNotFound) {
		t.Errorf("CopyCommand() err = %v, want ErrSSHCopyIDNotFound", err)
	}
}

func TestCopyCommand_Args(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sshDir, "mykey.pub"), []byte("pub"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := CopyKeyOpts{
		SSHDir:    sshDir,
		Key:       KeyEntry{RelPath: "mykey"},
		HostAlias: "homelab",
		User:      "gana",
	}
	cmd, err := CopyCommand(opts)
	if err != nil {
		t.Fatalf("CopyCommand: %v", err)
	}

	pubKeyPath := filepath.Join(sshDir, "mykey.pub")
	args := cmd.Args
	if len(args) < 4 {
		t.Fatalf("expected at least 4 args, got %v", args)
	}
	if args[1] != "-i" {
		t.Errorf("args[1] = %q, want -i", args[1])
	}
	if args[2] != pubKeyPath {
		t.Errorf("args[2] = %q, want %q", args[2], pubKeyPath)
	}
	if args[3] != "gana@homelab" {
		t.Errorf("args[3] = %q, want gana@homelab", args[3])
	}
}
