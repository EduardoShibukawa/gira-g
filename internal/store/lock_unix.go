//go:build unix || linux || darwin || freebsd || openbsd || netbsd

package store

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func platformLockFile(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}

func platformUnlockFile(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func isLockedError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)
}

func isUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, unix.ENOLCK) ||
		errors.Is(err, unix.ENOTSUP) ||
		errors.Is(err, unix.EOPNOTSUPP) ||
		errors.Is(err, unix.EROFS) ||
		errors.Is(err, unix.EACCES) ||
		errors.Is(err, unix.EPERM) ||
		errors.Is(err, unix.EINVAL)
}

func getPhysicalFileID(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("unable to read syscall.Stat_t from %s", path)
	}
	return fmt.Sprintf("%d-%d", stat.Dev, stat.Ino), nil
}
