package network

import (
	"os"
	"strings"
	"testing"
)

func TestHosts(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "hosts-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	hostsPathOverride = tmpFile.Name()
	defer func() { hostsPathOverride = "" }()

	t.Run("GetHostsPath", func(t *testing.T) {
		path := GetHostsPath()
		if path != hostsPathOverride {
			t.Errorf("GetHostsPath() = %v, want %v", path, hostsPathOverride)
		}
	})

	t.Run("Add Host", func(t *testing.T) {
		err := AddHost("127.0.0.1", "test.local")
		if err != nil {
			t.Fatalf("AddHost() error = %v", err)
		}

		content, _ := os.ReadFile(hostsPathOverride)
		if !strings.Contains(string(content), "test.local") {
			t.Errorf("Expected content to contain test.local")
		}

		// Add again should not duplicate or error
		err = AddHost("127.0.0.1", "test.local")
		if err != nil {
			t.Fatalf("AddHost() second call error = %v", err)
		}
	})

	t.Run("Add Host with different IP", func(t *testing.T) {
		// Currently AddHost returns nil if hostname exists regardless of IP
		err := AddHost("127.0.0.2", "test.local")
		if err != nil {
			t.Fatalf("AddHost() error = %v", err)
		}
	})

	t.Run("Remove Managed Hosts", func(t *testing.T) {
		err := RemoveManagedHosts()
		if err != nil {
			t.Fatalf("RemoveManagedHosts() error = %v", err)
		}

		content, _ := os.ReadFile(hostsPathOverride)
		if strings.Contains(string(content), "#WailsManaged") {
			t.Errorf("Expected content to NOT contain #WailsManaged")
		}
	})

	t.Run("AddHost - File not exists", func(t *testing.T) {
		hostsPathOverride = "/nonexistent/path/to/hosts"
		err := AddHost("127.0.0.1", "test.local")
		if err == nil {
			t.Errorf("AddHost should have failed for nonexistent file")
		}
	})

	t.Run("RemoveManagedHosts - File not exists", func(t *testing.T) {
		hostsPathOverride = "/nonexistent/path/to/hosts"
		err := RemoveManagedHosts()
		if err == nil {
			t.Errorf("RemoveManagedHosts should have failed for nonexistent file")
		}
	})
}
