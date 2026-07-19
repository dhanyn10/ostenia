package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"runtime"
	"strconv"
	"strings"
)

// MockExecutor implements utils.CommandExecutor for testing.
type MockExecutor struct {
	Output string
	Err    error
}

func (m *MockExecutor) Command(name string, arg ...string) *exec.Cmd {
	exitCode := 0
	if m.Err != nil {
		exitCode = 1
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if m.Output == "" {
			cmd = exec.Command("cmd", "/c", "exit", strconv.Itoa(exitCode)) // NOSONAR
		} else {
			// Powershell is used to ensure exact output without trailing newlines and correct exit code.
			// [Console]::Out.Write is used instead of Write-Host to avoid unwanted newlines.
			encodedOutput := strings.ReplaceAll(m.Output, "'", "''")
			script := fmt.Sprintf("[Console]::Out.Write('%s'); exit %d", encodedOutput, exitCode)
			cmd = exec.Command("powershell", "-NoProfile", "-Command", script) // NOSONAR
		}
	} else {
		if m.Output == "" {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("exit %d", exitCode)) // NOSONAR
		} else {
			// printf is used to ensure exact output without trailing newlines.
			encodedOutput := strings.ReplaceAll(m.Output, "'", "'\\''")
			script := fmt.Sprintf("printf '%%s' '%s'; exit %d", encodedOutput, exitCode)
			cmd = exec.Command("sh", "-c", script) // NOSONAR
		}
	}

	cmd.Env = utils.SafeEnv()
	return cmd
}

// MockHTTPClient implements utils.HTTPClient for testing.
type MockHTTPClient struct {
	Content string
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(m.Content)),
	}, nil
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(m.Content)),
	}, nil
}
