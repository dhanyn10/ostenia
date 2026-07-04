package php

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type mockExecutor struct {
	responses map[string]string
}

func (m *mockExecutor) Command(name string, arg ...string) *exec.Cmd {
	all := name + " " + strings.Join(arg, " ")

	var response string
	for k, v := range m.responses {
		if strings.Contains(all, k) {
			response = v
			break
		}
	}

	exitCode := "0"
	if response == "__ERROR__" {
		exitCode = "1"
		response = ""
	}

	argList := []string{"-test.run=TestHelperProcess", "--"}
	argList = append(argList, arg...)
	cmd := exec.Command(os.Args[0], argList...)
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"MOCK_OUTPUT="+response,
		"MOCK_EXIT_CODE="+exitCode,
	)
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("MOCK_OUTPUT") != "" {
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_OUTPUT"))
	}
	exitCode := 0
	fmt.Sscanf(os.Getenv("MOCK_EXIT_CODE"), "%d", &exitCode)
	os.Exit(exitCode)
}
