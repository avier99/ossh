package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	FlagWeakPerms        = "weak-perms"
	FlagNoPassphrase     = "no-passphrase"
	FlagLegacyKey        = "legacy-key"
	FlagHostKeyNotPinned = "host-key-not-pinned"
)

// Finding describes a single permissions or health issue under sshDir.
type Finding struct {
	Path        string
	Flag        string
	Description string
}

// Audit scans sshDir for permission issues. It never modifies anything.
func Audit(sshDir string) ([]Finding, error) {
	findings := make([]Finding, 0)

	info, err := os.Stat(sshDir)
	if err != nil {
		return nil, err
	}
	if perm := info.Mode().Perm(); perm != 0700 {
		findings = append(findings, Finding{
			Path:        sshDir,
			Flag:        FlagWeakPerms,
			Description: fmt.Sprintf("SSH directory should be mode 0700, found %04o", perm),
		})
	}

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(sshDir, name)

		var expected os.FileMode
		switch {
		case name == "config", name == "authorized_keys":
			expected = 0600
		case strings.HasSuffix(name, ".pub"):
			expected = 0644
		case name == "known_hosts", strings.HasSuffix(name, ".lock"):
			continue
		default:
			expected = 0600
		}

		fi, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if perm := fi.Mode().Perm(); perm != expected {
			findings = append(findings, Finding{
				Path:        path,
				Flag:        FlagWeakPerms,
				Description: fmt.Sprintf("expected mode %04o, found %04o", expected, perm),
			})
		}
	}

	return findings, nil
}
