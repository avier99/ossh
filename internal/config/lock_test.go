package config_test

import (
	"errors"
	"testing"

	"github.com/synoxisllc/ossh/internal/config"
)

func TestAcquireLockFreshDir(t *testing.T) {
	sshDir := t.TempDir()

	release, err := config.AcquireLock(sshDir)
	if err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}
	if release == nil {
		t.Fatal("release func is nil")
	}
	release()
}

func TestAcquireLockHeld(t *testing.T) {
	sshDir := t.TempDir()

	release, err := config.AcquireLock(sshDir)
	if err != nil {
		t.Fatalf("first AcquireLock: %v", err)
	}
	defer release()

	_, err = config.AcquireLock(sshDir)
	if !errors.Is(err, config.ErrLockHeld) {
		t.Fatalf("second AcquireLock err = %v, want ErrLockHeld", err)
	}
}

func TestAcquireLockReleaseAndReacquire(t *testing.T) {
	sshDir := t.TempDir()

	release, err := config.AcquireLock(sshDir)
	if err != nil {
		t.Fatalf("first AcquireLock: %v", err)
	}
	release()

	release, err = config.AcquireLock(sshDir)
	if err != nil {
		t.Fatalf("second AcquireLock: %v", err)
	}
	release()
}
