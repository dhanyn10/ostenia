//go:build !windows

package service

import "os/exec"

func notifyEnvironmentUpdate() {}

func SetHideWindow(cmd *exec.Cmd) {}
