package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/synoxisllc/ossh/internal/config"
)

func TestSetupFreshDir(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")

	if err := config.Setup(sshDir); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	info, err := os.Stat(sshDir)
	if err != nil {
		t.Fatalf("stat sshDir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("sshDir is not a directory")
	}
	if got := info.Mode().Perm(); got != 0700 {
		t.Errorf("sshDir mode = %o, want 0700", got)
	}

	configPath := filepath.Join(sshDir, "config")
	cfgInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := cfgInfo.Mode().Perm(); got != 0600 {
		t.Errorf("config mode = %o, want 0600", got)
	}
}

func TestSetupDirExistsNoConfig(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.Mkdir(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	if err := config.Setup(sshDir); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	configPath := filepath.Join(sshDir, "config")
	cfgInfo, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := cfgInfo.Mode().Perm(); got != 0600 {
		t.Errorf("config mode = %o, want 0600", got)
	}
}

func TestSetupBothExist(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.Mkdir(sshDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte("# existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := config.Setup(sshDir); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("config was modified:\nbefore: %q\nafter:  %q", before, after)
	}

	info, err := os.Stat(sshDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0755 {
		t.Errorf("sshDir mode changed to %o, want unchanged 0755", got)
	}
}
