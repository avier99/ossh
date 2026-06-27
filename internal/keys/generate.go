package keys

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KeyType identifies the algorithm used for key generation.
type KeyType string

const (
	KeyTypeED25519 KeyType = "ed25519"
	KeyTypeRSA4096 KeyType = "rsa4096"
)

// GenerateOpts holds the inputs for key generation.
type GenerateOpts struct {
	SSHDir  string
	SubDir  string // optional, e.g. "homelab-keys" — relative, one level only
	Name    string
	KeyType KeyType
}

var (
	ErrEmptyName         = errors.New("key name cannot be empty")
	ErrPathSeparator     = errors.New("key name cannot contain path separators")
	ErrSpaceInName       = errors.New("key name cannot contain spaces")
	ErrLeadingDot        = errors.New("key name cannot start with a dot")
	ErrInvalidSubDir     = errors.New("subfolder name cannot contain path separators or be absolute")
	ErrSSHKeygenNotFound = errors.New("ssh-keygen not found on PATH")
)

func targetDir(opts GenerateOpts) string {
	if opts.SubDir == "" {
		return opts.SSHDir
	}
	return filepath.Join(opts.SSHDir, opts.SubDir)
}

// ValidateNameFormat checks only the key name format rules, not file collision.
func ValidateNameFormat(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if strings.ContainsAny(name, `/\`) {
		return ErrPathSeparator
	}
	if strings.Contains(name, " ") {
		return ErrSpaceInName
	}
	if strings.HasPrefix(name, ".") {
		return ErrLeadingDot
	}
	return nil
}

// Validate checks that the key name is safe and target files do not already exist.
func Validate(opts GenerateOpts) error {
	if err := ValidateNameFormat(opts.Name); err != nil {
		return err
	}

	if opts.SubDir != "" {
		if strings.ContainsAny(opts.SubDir, `/\`) {
			return ErrInvalidSubDir
		}
		if strings.HasPrefix(opts.SubDir, ".") {
			return ErrInvalidSubDir
		}
		if filepath.IsAbs(opts.SubDir) {
			return ErrInvalidSubDir
		}
	}

	dir := targetDir(opts)
	privatePath := filepath.Join(dir, opts.Name)
	publicPath := privatePath + ".pub"

	if _, err := os.Stat(privatePath); err == nil {
		return fmt.Errorf("private key already exists: %s", privatePath)
	} else if !os.IsNotExist(err) {
		return err
	}

	if _, err := os.Stat(publicPath); err == nil {
		return fmt.Errorf("public key already exists: %s", publicPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Command builds an ssh-keygen invocation for interactive passphrase prompting.
// The caller decides how to run the returned *exec.Cmd (e.g. via tea.ExecProcess).
func Command(opts GenerateOpts) (*exec.Cmd, error) {
	if err := Validate(opts); err != nil {
		return nil, err
	}

	sshKeygen, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return nil, ErrSSHKeygenNotFound
	}

	dir := targetDir(opts)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	keyPath := filepath.Join(dir, opts.Name)

	var args []string
	switch opts.KeyType {
	case KeyTypeED25519:
		args = []string{"-t", "ed25519", "-f", keyPath, "-C", ""}
	case KeyTypeRSA4096:
		args = []string{"-t", "rsa", "-b", "4096", "-f", keyPath, "-C", ""}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", opts.KeyType)
	}

	cmd := exec.Command(sshKeygen, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, nil
}

// Verify checks that generated key files exist with correct permissions.
func Verify(sshDir, name string) error {
	privatePath := filepath.Join(sshDir, name)
	publicPath := privatePath + ".pub"

	privateInfo, err := os.Stat(privatePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("private key not found: %s", privatePath)
	} else if err != nil {
		return err
	}
	if privateInfo.Mode().Perm() != 0600 {
		return fmt.Errorf("private key %s has permissions %04o, expected 0600", privatePath, privateInfo.Mode().Perm())
	}

	publicInfo, err := os.Stat(publicPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("public key not found: %s", publicPath)
	} else if err != nil {
		return err
	}
	if publicInfo.Mode().Perm() != 0644 {
		return fmt.Errorf("public key %s has permissions %04o, expected 0644", publicPath, publicInfo.Mode().Perm())
	}

	return nil
}

// HasPassphrase reports whether the private key at sshDir/name is protected by a passphrase.
//
// HasPassphrase requires an actual key on disk — it cannot be meaningfully unit tested.
// Same pattern as ssh.Connect and syscall.Exec.
func HasPassphrase(sshDir, name string) (bool, error) {
	sshKeygen, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return false, ErrSSHKeygenNotFound
	}

	keyPath := filepath.Join(sshDir, name)
	cmd := exec.Command(sshKeygen, "-y", "-P", "", "-f", keyPath)
	cmd.Stderr = io.Discard

	err = cmd.Run()
	if err == nil {
		return false, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return true, nil
	}

	return false, err
}
