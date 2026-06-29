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
}
