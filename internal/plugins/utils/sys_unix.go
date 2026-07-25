//go:build !windows

package utils

import (
	"os/exec"
)

func setHideWindow(cmd *exec.Cmd) {
	// This function is a no-op on non-Windows/Unix-based platforms because
	// console window hiding (SysProcAttr.HideWindow) is not supported
	// or necessary outside of the Windows operating system environment.
}

// SetHideWindow is a no-op on non-Windows platforms.
func SetHideWindow(cmd *exec.Cmd) {
	setHideWindow(cmd)
}
