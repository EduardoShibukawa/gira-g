//go:build !windows

package windowsgates

import (
	"errors"
	"os"
	"syscall"
)

// tryLockExclusive takes flock(LOCK_EX|LOCK_NB). (false, nil) means contended.
func tryLockExclusive(f *os.File) (bool, error) {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return false, nil
	}
	return false, err
}

func unlock(f *os.File) { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }
