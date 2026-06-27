package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupPrefix = "config.ossh-backup-"

// Backup copies ~/.ssh/config to a timestamped backup file. If config does not
// exist, Backup is a no-op.
func Backup(sshDir string) error {
	configPath := filepath.Join(sshDir, "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	ts := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(sshDir, backupPrefix+ts)
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return err
	}

	return pruneBackups(sshDir)
}

func pruneBackups(sshDir string) error {
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return err
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, backupPrefix) {
			backups = append(backups, name)
		}
	}

	sort.Strings(backups)

	if len(backups) <= 5 {
		return nil
	}

	for _, name := range backups[:len(backups)-5] {
		_ = os.Remove(filepath.Join(sshDir, name))
	}

	return nil
}
