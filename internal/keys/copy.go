package keys

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var ErrSSHCopyIDNotFound = errors.New("ssh-copy-id not found on PATH")

// CopyKeyOpts holds inputs for copying a public key to a remote host.
type CopyKeyOpts struct {
	SSHDir    string
	Key       KeyEntry
	HostAlias string
	User      string // optional
}

// Target returns the ssh-copy-id destination (user@alias or alias).
func Target(opts CopyKeyOpts) string {
	if opts.User != "" {
		return opts.User + "@" + opts.HostAlias
	}
	return opts.HostAlias
}

// CopyCommand builds an ssh-copy-id command for the given options.
func CopyCommand(opts CopyKeyOpts) (*exec.Cmd, error) {
	path, err := exec.LookPath("ssh-copy-id")
	if err != nil {
		return nil, ErrSSHCopyIDNotFound
	}
	pubKeyPath := filepath.Join(opts.SSHDir, opts.Key.RelPath) + ".pub"
	cmd := exec.Command(path, "-i", pubKeyPath, Target(opts))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

// ManualFallback returns a shell one-liner when ssh-copy-id is unavailable or fails.
func ManualFallback(opts CopyKeyOpts) string {
	pubKey := opts.Key.DisplayPath + ".pub"
	target := Target(opts)
	return fmt.Sprintf(`cat %s | ssh %s "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"`, pubKey, target)
}
