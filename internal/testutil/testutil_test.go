package testutil

import "testing"

func TestMockExecutor(t *testing.T) {
	m := &MockExecutor{Output: "success"}
	cmd := m.Command("echo", "test")
	if cmd == nil {
		t.Fatal("Expected non-nil command")
	}
}

func TestMockHTTPClient(t *testing.T) {
	m := &MockHTTPClient{Content: "{\"status\":\"ok\"}"}
	resp, err := m.Get("http://example.com")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}
