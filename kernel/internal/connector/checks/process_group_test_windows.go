//go:build windows

package checks

import "syscall"

func processGroupAttributes() *syscall.SysProcAttr { return &syscall.SysProcAttr{} }

func terminateProcessGroup(pid int) error { return syscall.Errno(0) }
