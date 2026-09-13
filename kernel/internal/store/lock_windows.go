//go:build windows

package store

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	lockFileFailImmediately = 0x00000001
	lockFileExclusive       = 0x00000002
	errorLockViolation      = syscall.Errno(33)
)

var (
	errLockBusy = syscall.EWOULDBLOCK
	kernel32    = syscall.NewLazyDLL("kernel32.dll")
	lockFileEx  = kernel32.NewProc("LockFileEx")
)

func lockPlatform(f *os.File) error {
	var overlapped syscall.Overlapped
	result, _, err := lockFileEx.Call(
		uintptr(f.Fd()),
		lockFileFailImmediately|lockFileExclusive,
		0,
		1,
		0,
		uintptr(unsafe.Pointer(&overlapped)),
	)
	if result != 0 {
		return nil
	}
	if err == errorLockViolation {
		return errLockBusy
	}
	return err
}
