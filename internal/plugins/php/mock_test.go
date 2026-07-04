package php

import (
	"os/exec"
	"strings"
)

type mockExecutor struct {
	responses map[string]string
}

func (m *mockExecutor) Command(name string, arg ...string) *exec.Cmd {
	// join all to check
	all := name + " " + strings.Join(arg, " ")

	var response string
	// Check if any of our keys is in the full command
	for k, v := range m.responses {
		if strings.Contains(all, k) {
			response = v
			break
		}
	}

	if response == "__ERROR__" {
		return exec.Command("false")
	}

	return exec.Command("echo", "-n", response)
}
