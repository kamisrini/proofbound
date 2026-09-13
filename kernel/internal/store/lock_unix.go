//go:build !windows

package store

import (
	"os"
	"syscall"
)

var errLockBusy = syscall.EWOULDBLOCK

func lockPlatform(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}
