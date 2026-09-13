//go:build !windows

package checks

import "syscall"

func processGroupAttributes() *syscall.SysProcAttr { return &syscall.SysProcAttr{Setpgid: true} }

func terminateProcessGroup(pid int) error { return syscall.Kill(-pid, syscall.SIGTERM) }
