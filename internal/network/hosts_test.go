package network

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHosts(t *testing.T) {
	// Buat file hosts sementara
	tmpDir, err := os.MkdirTemp("", "hosts_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hostsPath := filepath.Join(tmpDir, "hosts")
	initialContent := "127.0.0.1 localhost\n"
	err = os.WriteFile(hostsPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("failed to write temp hosts: %v", err)
	}

	// Override hosts path
	hostsPathOverride = hostsPath
	defer func() { hostsPathOverride = "" }()

	// Test AddHost
	err = AddHost("127.0.0.1", "test.local")
	if err != nil {
		t.Errorf("AddHost failed: %v", err)
	}

	content, _ := os.ReadFile(hostsPath)
	if !strings.Contains(string(content), "test.local") {
		t.Errorf("Expected content to contain test.local, got: %s", string(content))
	}

	// Test AddHost already exists
	err = AddHost("127.0.0.1", "test.local")
	if err != nil {
		t.Errorf("AddHost already exists failed: %v", err)
	}

	// Test RemoveManagedHosts
	err = RemoveManagedHosts()
	if err != nil {
		t.Errorf("RemoveManagedHosts failed: %v", err)
	}

	content, _ = os.ReadFile(hostsPath)
	if strings.Contains(string(content), "test.local") {
		t.Errorf("Expected content NOT to contain test.local after removal, got: %s", string(content))
	}
}
