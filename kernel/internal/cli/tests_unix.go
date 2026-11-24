//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
)

// killGroupOnCancel starts the tests in their own process group and kills the whole group on
// cancel or timeout, so no process the tests started keeps running (tests-can-fail review F4).
func killGroupOnCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
}
