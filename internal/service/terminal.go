package service

import (
	"os/exec"
	"runtime"
)

type Terminal struct {
	WorkingDir string
	Env        []string
}

func NewTerminal(workingDir string, env []string) *Terminal {
	return &Terminal{
		WorkingDir: workingDir,
		Env:        env,
	}
}

func (t *Terminal) Start() error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/K", "title Ostenia Terminal")
	} else {
		// Mock for Linux
		cmd = exec.Command("bash")
	}

	cmd.Dir = t.WorkingDir
	cmd.Env = t.Env

	// For Windows, we want it to open in a new window
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", "title Ostenia Terminal")
		cmd.Dir = t.WorkingDir
		cmd.Env = t.Env
	}

	return cmd.Start()
}
