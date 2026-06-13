package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/synoxisllc/ossh/internal/config"
)

func TestAuditAllCorrect(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.Chmod(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := config.Audit(sshDir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d: %+v", len(findings), findings)
	}
}

func TestAuditWeakSSHDir(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.Chmod(sshDir, 0755); err != nil {
		t.Fatal(err)
	}

	findings, err := config.Audit(sshDir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Path != sshDir {
		t.Errorf("finding path = %q, want %q", f.Path, sshDir)
	}
	if f.Flag != config.FlagWeakPerms {
		t.Errorf("finding flag = %q, want %q", f.Flag, config.FlagWeakPerms)
	}
}

func TestAuditWeakPrivateKey(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.Chmod(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(sshDir, "id_ed25519")
	if err := os.WriteFile(keyPath, []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := config.Audit(sshDir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Path != keyPath {
		t.Errorf("finding path = %q, want %q", f.Path, keyPath)
	}
	if f.Flag != config.FlagWeakPerms {
		t.Errorf("finding flag = %q, want %q", f.Flag, config.FlagWeakPerms)
	}
}

func TestAuditWeakAuthorizedKeys(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.Chmod(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	authKeysPath := filepath.Join(sshDir, "authorized_keys")
	if err := os.WriteFile(authKeysPath, []byte("ssh-ed25519 AAAA"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := config.Audit(sshDir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Path != authKeysPath {
		t.Errorf("finding path = %q, want %q", f.Path, authKeysPath)
	}
	if f.Flag != config.FlagWeakPerms {
		t.Errorf("finding flag = %q, want %q", f.Flag, config.FlagWeakPerms)
	}
}

func TestAuditMultipleIssues(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.Chmod(sshDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(sshDir, "id_ed25519")
	if err := os.WriteFile(keyPath, []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := config.Audit(sshDir)
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d: %+v", len(findings), findings)
	}

	paths := map[string]bool{}
	for _, f := range findings {
		if f.Flag != config.FlagWeakPerms {
			t.Errorf("finding for %q has flag %q, want %q", f.Path, f.Flag, config.FlagWeakPerms)
		}
		paths[f.Path] = true
	}
	for _, want := range []string{sshDir, configPath, keyPath} {
		if !paths[want] {
			t.Errorf("missing finding for %q", want)
		}
	}
}
