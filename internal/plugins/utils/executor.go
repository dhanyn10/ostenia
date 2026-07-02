package utils

import (
	"os/exec"
)

// CommandExecutor is an interface for executing system commands, allowing for test mocking.
type CommandExecutor interface {
	Command(name string, arg ...string) *exec.Cmd
}

// DefaultExecutor is the production implementation of CommandExecutor using os/exec.
type DefaultExecutor struct{}

func (e *DefaultExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// Executor is the global command runner instance.
var Executor CommandExecutor = &DefaultExecutor{}
