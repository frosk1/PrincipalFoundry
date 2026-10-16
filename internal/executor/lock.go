package executor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// ErrAlreadyRunning is returned when another live executor cycle holds the lock.
var ErrAlreadyRunning = errors.New("executor cycle already running")

// lockFileName is stored inside the git directory so it is never committed.
const lockFileName = "pfoundry-executor.lock"

// Lock is a local, non-committed executor lock stored inside the git directory.
// It is implemented with an advisory flock on an open file descriptor, so it is
// released automatically if the process crashes, is killed, or the machine
// reboots. Overlapping cycles therefore cannot deadlock on a leftover file.
type Lock struct {
	file *os.File
}

// Acquire takes the executor lock for the repository whose git directory is
// gitDir. It returns ErrAlreadyRunning when another process currently holds it.
func Acquire(gitDir string) (*Lock, error) {
	path := filepath.Join(gitDir, lockFileName)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrAlreadyRunning
		}
		return nil, err
	}

	// Record the owning pid for diagnostics. Contents are not used for locking.
	if err := file.Truncate(0); err == nil {
		if _, err := file.Seek(0, 0); err == nil {
			fmt.Fprintf(file, "%d\n", os.Getpid())
		}
	}
	return &Lock{file: file}, nil
}

// Held reports whether another process currently holds the executor lock. It
// does not create the lock file and does not modify repository state.
func Held(gitDir string) bool {
	file, err := os.Open(filepath.Join(gitDir, lockFileName))
	if err != nil {
		return false
	}
	defer file.Close()

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return errors.Is(err, syscall.EWOULDBLOCK)
	}
	syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	return false
}

// Release unlocks and closes the lock. The lock file itself is left in place,
// which is harmless and avoids racing with a concurrent Held check.
func (l *Lock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	err := l.file.Close()
	l.file = nil
	return err
}
