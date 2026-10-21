package testutil

import (
	"errors"
	"net/http"
	"testing"
)

func TestMockExecutor(t *testing.T) {
	t.Run("with output and success", func(t *testing.T) {
		m := &MockExecutor{Output: "success"}
		cmd := m.Command("echo", "test")
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command execution failed: %v", err)
		}
		if string(out) != "success" {
			t.Errorf("Expected 'success', got %q", string(out))
		}
	})

	t.Run("empty output with error", func(t *testing.T) {
		m := &MockExecutor{Err: errors.New("command failed")}
		cmd := m.Command("some-cmd")
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}
		err := cmd.Run()
		if err == nil {
			t.Fatal("Expected command to exit with error")
		}
	})

	t.Run("with output and error", func(t *testing.T) {
		m := &MockExecutor{Output: "failure output", Err: errors.New("fail")}
		cmd := m.Command("fail-cmd")
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatal("Expected error from command")
		}
		if string(out) != "failure output" {
			t.Errorf("Expected 'failure output', got %q", string(out))
		}
	})

	t.Run("empty output and success", func(t *testing.T) {
		m := &MockExecutor{}
		cmd := m.Command("dummy")
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}
		err := cmd.Run()
		if err != nil {
			t.Fatalf("Expected command to succeed, got %v", err)
		}
	})
}

func TestMockHTTPClient(t *testing.T) {
	t.Run("Get method", func(t *testing.T) {
		m := &MockHTTPClient{Content: `{"status":"ok"}`}
		resp, err := m.Get("http://example.com")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Do method", func(t *testing.T) {
		m := &MockHTTPClient{Content: `hello world`}
		req, err := http.NewRequest(http.MethodGet, "http://example.com/test", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := m.Do(req)
		if err != nil {
			t.Fatalf("Do failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}
