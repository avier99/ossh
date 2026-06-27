package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackup_NoConfigIsNoOp(t *testing.T) {
	sshDir := t.TempDir()

	if err := Backup(sshDir); err != nil {
		t.Fatalf("Backup: %v", err)
	}

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), backupPrefix) {
			t.Errorf("unexpected backup file: %s", e.Name())
		}
	}
}

func TestBackup_CreatesBackupWithCorrectContent(t *testing.T) {
	sshDir := t.TempDir()
	original := "Host test\n    HostName 1.2.3.4\n"
	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	if err := Backup(sshDir); err != nil {
		t.Fatalf("Backup: %v", err)
	}

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		t.Fatal(err)
	}

	var backupPath string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), backupPrefix) {
			backupPath = filepath.Join(sshDir, e.Name())
			break
		}
	}
	if backupPath == "" {
		t.Fatal("no backup file created")
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Errorf("backup content = %q, want %q", data, original)
	}

	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("backup mode = %o, want 0600", got)
	}
}

func TestBackup_PrunesTo5(t *testing.T) {
	sshDir := t.TempDir()

	names := []string{
		"config.ossh-backup-20260101-000001",
		"config.ossh-backup-20260102-000001",
		"config.ossh-backup-20260103-000001",
		"config.ossh-backup-20260104-000001",
		"config.ossh-backup-20260105-000001",
		"config.ossh-backup-20260106-000001",
		"config.ossh-backup-20260107-000001",
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(sshDir, name), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	if err := pruneBackups(sshDir); err != nil {
		t.Fatalf("pruneBackups: %v", err)
	}

	remaining := make(map[string]bool)
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), backupPrefix) {
			remaining[e.Name()] = true
		}
	}

	if len(remaining) != 5 {
		t.Fatalf("expected 5 backups, got %d: %v", len(remaining), remaining)
	}

	want := []string{
		"config.ossh-backup-20260103-000001",
		"config.ossh-backup-20260104-000001",
		"config.ossh-backup-20260105-000001",
		"config.ossh-backup-20260106-000001",
		"config.ossh-backup-20260107-000001",
	}
	for _, name := range want {
		if !remaining[name] {
			t.Errorf("expected backup %q to remain", name)
		}
	}
}
