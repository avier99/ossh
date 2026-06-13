package ssh

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

var ErrEmptyHostAlias = errors.New("host alias cannot be empty")
var ErrSSHNotFound = errors.New("ssh binary not found on PATH")

// Never returns on success — the process is replaced.
func Connect(hostAlias string) error {
	if hostAlias == "" {
		return ErrEmptyHostAlias
	}

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return ErrSSHNotFound
	}

	// syscall.Exec replaces the current process with ssh.
	// On success, this never returns.
	// argv[0] is the program name as seen by ssh.
	err = syscall.Exec(sshPath, []string{"ssh", hostAlias}, os.Environ())
	
	// If we reach here, exec failed
	return err
}
