package config

import (
	"os"

	"github.com/kevinburke/ssh_config"
)

// Load reads and parses an SSH config file at configPath.
func Load(configPath string) (*ssh_config.Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ssh_config.Decode(f)
}
