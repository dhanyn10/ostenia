//go:build !windows

package utils

import (
	"os/exec"
)

func setHideWindow(cmd *exec.Cmd) {}

// SetHideWindow is a no-op on non-Windows platforms.
func SetHideWindow(cmd *exec.Cmd) {
	setHideWindow(cmd)
}
