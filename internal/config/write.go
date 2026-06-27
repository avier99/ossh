package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
)

var (
	ErrEmptyAlias      = errors.New("alias cannot be empty")
	ErrSpaceInAlias    = errors.New("alias cannot contain spaces")
	ErrWildcardInAlias = errors.New("alias cannot contain wildcard characters")
)

// AddHostOpts holds the fields for a new ossh-managed Host block.
type AddHostOpts struct {
	Alias        string
	Hostname     string
	User         string // omit directive if blank
	Port         int    // omit directive if 0 or 22
	AuthMethod   string // "key" or "password"
	IdentityFile string // omit directive if blank
}

// ValidateAlias checks alias format and uniqueness against cfg.
func ValidateAlias(alias string, cfg *ssh_config.Config) error {
	if alias == "" {
		return ErrEmptyAlias
	}
	if strings.ContainsAny(alias, " \t") {
		return ErrSpaceInAlias
	}
	if strings.ContainsAny(alias, "*?") {
		return ErrWildcardInAlias
	}
	if cfg != nil {
		for _, host := range cfg.Hosts {
			for _, p := range host.Patterns {
				if p.String() == alias {
					return fmt.Errorf("alias %q already exists in config", alias)
				}
			}
		}
	}
	return nil
}

// AddHost appends an ossh-managed Host block to sshDir/config.
func AddHost(sshDir string, opts AddHostOpts) error {
	if err := Backup(sshDir); err != nil {
		return err
	}

	block := buildHostBlock(opts)
	configPath := filepath.Join(sshDir, "config")

	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n%s\n", block); err != nil {
		return err
	}

	return nil
}

func buildHostBlock(opts AddHostOpts) string {
	var b strings.Builder
	b.WriteString("# ossh-managed: " + opts.Alias + "\n")
	b.WriteString("Host " + opts.Alias + "\n")
	b.WriteString("\tHostName " + opts.Hostname + "\n")
	if opts.User != "" {
		b.WriteString("\tUser " + opts.User + "\n")
	}
	if opts.Port != 0 && opts.Port != 22 {
		b.WriteString(fmt.Sprintf("\tPort %d\n", opts.Port))
	}
	if opts.AuthMethod == "key" && opts.IdentityFile != "" {
		b.WriteString("\tIdentityFile " + opts.IdentityFile + "\n")
	}
	if opts.AuthMethod == "password" {
		b.WriteString("\tPasswordAuthentication yes\n")
	}
	b.WriteString("# /ossh-managed")
	return b.String()
}
