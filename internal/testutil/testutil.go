package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// MockExecutor implements utils.CommandExecutor for testing.
type MockExecutor struct {
	Output string
	Err    error
}

func (m *MockExecutor) Command(name string, arg ...string) *exec.Cmd {
	// Special Case: On Windows, re-executing os.Args[0] (the test binary)
	// causes file locking that leads to "Access is denied" when the test runner
	// tries to clean up the temporary directory.

	// If we don't need output or error code 1, we can use a system command
	// that doesn't lock our own binary.
	if m.Output == "" && m.Err == nil {
		if runtime.GOOS == "windows" {
			return exec.Command("cmd", "/c", "exit", "0")
		}
		return exec.Command("true")
	}

	argList := []string{"-test.run=TestHelperProcess", "--", name}
	argList = append(argList, arg...)
	cmd := exec.Command(os.Args[0], argList...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "MOCK_OUTPUT="+m.Output)
	if m.Err != nil {
		cmd.Env = append(cmd.Env, "MOCK_EXIT_CODE=1")
	}
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

// HelperProcess is a generic test helper process.
func HelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("MOCK_OUTPUT") != "" {
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_OUTPUT"))
	}
	if os.Getenv("MOCK_EXIT_CODE") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
}
