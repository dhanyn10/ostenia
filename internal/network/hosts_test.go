package network

import (
	"os"
	"strings"
	"testing"
)

func setupHostsTempFile(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "hosts-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name()
}

func TestGetHostsPath(t *testing.T) {
	hostsPathOverride = "dummy-path"
	defer func() { hostsPathOverride = "" }()

	path := GetHostsPath()
	if path != hostsPathOverride {
		t.Errorf("GetHostsPath() = %v, want %v", path, hostsPathOverride)
	}
}

func TestAddHost(t *testing.T) {
	tempName := setupHostsTempFile(t)
	defer os.Remove(tempName)

	hostsPathOverride = tempName
	defer func() { hostsPathOverride = "" }()

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
}

func TestAddHostWithDifferentIP(t *testing.T) {
	tempName := setupHostsTempFile(t)
	defer os.Remove(tempName)

	hostsPathOverride = tempName
	defer func() { hostsPathOverride = "" }()

	err := AddHost("127.0.0.1", "test.local")
	if err != nil {
		t.Fatalf("AddHost() error = %v", err)
	}

	// Currently AddHost returns nil if hostname exists regardless of IP
	err = AddHost("127.0.0.2", "test.local")
	if err != nil {
		t.Fatalf("AddHost() error = %v", err)
	}
}

func TestRemoveManagedHosts(t *testing.T) {
	tempName := setupHostsTempFile(t)
	defer os.Remove(tempName)

	hostsPathOverride = tempName
	defer func() { hostsPathOverride = "" }()

	err := AddHost("127.0.0.1", "test.local")
	if err != nil {
		t.Fatalf("AddHost() error = %v", err)
	}

	err = RemoveManagedHosts()
	if err != nil {
		t.Fatalf("RemoveManagedHosts() error = %v", err)
	}

	content, _ := os.ReadFile(hostsPathOverride)
	if strings.Contains(string(content), "#WailsManaged") {
		t.Errorf("Expected content to NOT contain #WailsManaged")
	}
}

func TestAddHost_FileNotExists(t *testing.T) {
	hostsPathOverride = "/nonexistent/path/to/hosts"
	defer func() { hostsPathOverride = "" }()

	err := AddHost("127.0.0.1", "test.local")
	if err == nil {
		t.Errorf("AddHost should have failed for nonexistent file")
	}
}

func TestRemoveManagedHosts_FileNotExists(t *testing.T) {
	hostsPathOverride = "/nonexistent/path/to/hosts"
	defer func() { hostsPathOverride = "" }()

	err := RemoveManagedHosts()
	if err == nil {
		t.Errorf("RemoveManagedHosts should have failed for nonexistent file")
	}
}
