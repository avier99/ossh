package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildHostBlock_KeyAuth(t *testing.T) {
	block := buildHostBlock(AddHostOpts{
		Alias:        "homelab",
		Hostname:     "192.168.1.100",
		User:         "gana",
		AuthMethod:   "key",
		IdentityFile: "~/.ssh/id_ed25519",
	})

	if !strings.Contains(block, "# ossh-managed: homelab") {
		t.Error("missing ossh-managed marker")
	}
	if !strings.Contains(block, "IdentityFile ~/.ssh/id_ed25519") {
		t.Error("missing IdentityFile")
	}
	if strings.Contains(block, "PasswordAuthentication") {
		t.Error("should not contain PasswordAuthentication for key auth")
	}
}

func TestBuildHostBlock_PasswordAuth(t *testing.T) {
	block := buildHostBlock(AddHostOpts{
		Alias:      "legacy",
		Hostname:   "10.0.0.1",
		AuthMethod: "password",
	})

	if !strings.Contains(block, "PasswordAuthentication yes") {
		t.Error("missing PasswordAuthentication yes")
	}
	if strings.Contains(block, "IdentityFile") {
		t.Error("should not contain IdentityFile for password auth")
	}
}

func TestBuildHostBlock_OmitsPort22(t *testing.T) {
	block := buildHostBlock(AddHostOpts{
		Alias:      "homelab",
		Hostname:   "192.168.1.100",
		Port:       22,
		AuthMethod: "key",
	})

	if strings.Contains(block, "Port") {
		t.Error("Port 22 should be omitted")
	}
}

func TestBuildHostBlock_WritesNonStandardPort(t *testing.T) {
	block := buildHostBlock(AddHostOpts{
		Alias:      "homelab",
		Hostname:   "192.168.1.100",
		Port:       2222,
		AuthMethod: "key",
	})

	if !strings.Contains(block, "Port 2222") {
		t.Error("expected Port 2222 in block")
	}
}

func TestBuildHostBlock_OmitsBlankUser(t *testing.T) {
	block := buildHostBlock(AddHostOpts{
		Alias:      "homelab",
		Hostname:   "192.168.1.100",
		AuthMethod: "key",
	})

	if strings.Contains(block, "User") {
		t.Error("blank user should be omitted")
	}
}

func TestAddHost_AppendsBlock(t *testing.T) {
	sshDir := t.TempDir()
	original := "Host existing\n\tHostName 1.1.1.1\n"
	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	opts := AddHostOpts{
		Alias:        "newhost",
		Hostname:     "2.2.2.2",
		User:         "admin",
		AuthMethod:   "key",
		IdentityFile: "~/.ssh/mykey",
	}
	if err := AddHost(sshDir, opts); err != nil {
		t.Fatalf("AddHost: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.HasPrefix(content, original) {
		t.Error("original content was not preserved at start of file")
	}
	if !strings.Contains(content, "# ossh-managed: newhost") {
		t.Error("new block not appended")
	}
}

func TestAddHost_CreatesConfigIfMissing(t *testing.T) {
	sshDir := t.TempDir()

	opts := AddHostOpts{
		Alias:      "first",
		Hostname:   "10.0.0.1",
		AuthMethod: "password",
	}
	if err := AddHost(sshDir, opts); err != nil {
		t.Fatalf("AddHost: %v", err)
	}

	configPath := filepath.Join(sshDir, "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Host first") {
		t.Errorf("config missing host block: %q", data)
	}
}

func TestValidateAlias_Empty(t *testing.T) {
	err := ValidateAlias("", nil)
	if !errors.Is(err, ErrEmptyAlias) {
		t.Errorf("expected ErrEmptyAlias, got %v", err)
	}
}

func TestValidateAlias_Space(t *testing.T) {
	err := ValidateAlias("my host", nil)
	if !errors.Is(err, ErrSpaceInAlias) {
		t.Errorf("expected ErrSpaceInAlias, got %v", err)
	}
}

func TestValidateAlias_Wildcard(t *testing.T) {
	for _, alias := range []string{"host*", "host?"} {
		err := ValidateAlias(alias, nil)
		if !errors.Is(err, ErrWildcardInAlias) {
			t.Errorf("alias %q: expected ErrWildcardInAlias, got %v", alias, err)
		}
	}
}

func TestValidateAlias_AlreadyExists(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "sample.ssh_config"))
	if err != nil {
		t.Fatal(err)
	}

	err = ValidateAlias("home", cfg)
	if err == nil {
		t.Fatal("expected error for existing alias")
	}
	if !strings.Contains(err.Error(), `alias "home" already exists`) {
		t.Errorf("unexpected error: %v", err)
	}
}
