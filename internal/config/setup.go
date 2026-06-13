package config

import (
	"os"
	"path/filepath"
)

// Setup ensures sshDir exists with mode 0700 and an empty config file at mode 0600.
// Existing directories and files are left unchanged.
func Setup(sshDir string) error {
	info, err := os.Stat(sshDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(sshDir, 0700); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return &os.PathError{Op: "setup", Path: sshDir, Err: os.ErrInvalid}
	}

	configPath := filepath.Join(sshDir, "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		f, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		return f.Close()
	} else if err != nil {
		return err
	}

	return nil
}
