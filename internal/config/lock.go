package config

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// ErrLockHeld is returned when another ossh instance holds the config lock.
var ErrLockHeld = errors.New("config lock held by another ossh instance")

// AcquireLock opens sshDir/config.ossh.lock and acquires an exclusive non-blocking flock.
// On success it returns a release function that unlocks and closes the lock file.
func AcquireLock(sshDir string) (func(), error) {
	lockPath := filepath.Join(sshDir, "config.ossh.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrLockHeld
		}
		return nil, err
	}

	release := func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}
	return release, nil
}
