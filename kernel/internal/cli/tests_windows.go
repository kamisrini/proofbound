//go:build windows

package cli

import "os/exec"

// killGroupOnCancel: on Windows, exec's default cancel kills the shell only.
func killGroupOnCancel(cmd *exec.Cmd) {}
