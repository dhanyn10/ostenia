//go:build !windows

package utils

import (
	"os/exec"
)

// SetHideWindow is a no-op on non-Windows platforms.
func SetHideWindow(cmd *exec.Cmd) {}
