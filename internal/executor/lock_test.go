package executor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireAndRelease(t *testing.T) {
	gitDir := t.TempDir()

	lock, err := Acquire(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if !Held(gitDir) {
		t.Fatal("expected the lock to be held")
	}

	if _, err := Acquire(gitDir); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("expected ErrAlreadyRunning, got %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
	if Held(gitDir) {
		t.Fatal("expected the lock to be released")
	}

	// The lock can be re-acquired after release.
	again, err := Acquire(gitDir)
	if err != nil {
		t.Fatalf("expected to re-acquire the released lock, got %v", err)
	}
	again.Release()
}

func TestAcquireIgnoresAbandonedLockFile(t *testing.T) {
	gitDir := t.TempDir()
	// A lock file left behind without a live flock (for example after a crash)
	// must not be treated as held.
	if err := os.WriteFile(filepath.Join(gitDir, lockFileName), []byte("999999\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Held(gitDir) {
		t.Fatal("an abandoned lock file should not count as held")
	}

	lock, err := Acquire(gitDir)
	if err != nil {
		t.Fatalf("expected an abandoned lock file to be reusable, got %v", err)
	}
	lock.Release()
}

func TestHeldFalseWithoutLock(t *testing.T) {
	if Held(t.TempDir()) {
		t.Fatal("no lock should be reported as held")
	}
}
