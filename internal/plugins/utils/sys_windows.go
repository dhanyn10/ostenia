//go:build windows

package utils

import (
	"os/exec"
	"syscall"
)

// SetHideWindow sets the HideWindow attribute on Windows.
func SetHideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
