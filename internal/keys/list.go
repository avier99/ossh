package keys

import (
	"os"
	"path/filepath"
	"strings"
)

// KeyEntry describes a private key file under ~/.ssh.
type KeyEntry struct {
	RelPath     string // "mykey" or "homelab-keys/mykey"
	DisplayPath string // "~/.ssh/mykey" or "~/.ssh/homelab-keys/mykey"
}

// ListPrivateKeys returns private key files in sshDir (one subdirectory level deep).
func ListPrivateKeys(sshDir string) ([]KeyEntry, error) {
	homeDir, _ := os.UserHomeDir()

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, err
	}

	var keys []KeyEntry
	for _, entry := range entries {
		if entry.IsDir() {
			subEntries, err := os.ReadDir(filepath.Join(sshDir, entry.Name()))
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				if sub.IsDir() {
					continue
				}
				name := sub.Name()
				if !isPrivateKey(name) {
					continue
				}
				relPath := entry.Name() + "/" + name
				keys = append(keys, KeyEntry{
					RelPath:     relPath,
					DisplayPath: displayPath(homeDir, sshDir, relPath),
				})
			}
			continue
		}

		name := entry.Name()
		if !isPrivateKey(name) {
			continue
		}
		keys = append(keys, KeyEntry{
			RelPath:     name,
			DisplayPath: displayPath(homeDir, sshDir, name),
		})
	}

	return keys, nil
}

func displayPath(homeDir, sshDir, relPath string) string {
	if homeDir != "" && sshDir == filepath.Join(homeDir, ".ssh") {
		return "~/.ssh/" + relPath
	}
	return filepath.Join(sshDir, relPath)
}

func isPrivateKey(name string) bool {
	switch name {
	case "config", "known_hosts", "authorized_keys":
		return false
	}
	if strings.HasSuffix(name, ".pub") {
		return false
	}
	if strings.HasSuffix(name, ".lock") {
		return false
	}
	if strings.HasPrefix(name, "config.ossh-backup-") {
		return false
	}
	return true
}
