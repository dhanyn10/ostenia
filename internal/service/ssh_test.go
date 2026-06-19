package service

import (
	"os"
	"ostenia/internal/config"
	"path/filepath"
	"testing"
)

func TestGetAuth(t *testing.T) {
	m := &SSHManager{}

	t.Run("Password Auth", func(t *testing.T) {
		session := config.SSHSession{
			AuthMethod: "password",
			Password:   "testpass",
		}
		auth, err := m.getAuth(session)
		if err != nil {
			t.Fatalf("getAuth failed: %v", err)
		}
		if auth == nil {
			t.Fatal("expected non-nil auth")
		}
		// Since goph.Auth is a slice of ssh.AuthMethod, we just check if it's not empty
		if len(auth) == 0 {
			t.Error("expected auth methods to be present")
		}
	})

	t.Run("Key Auth", func(t *testing.T) {
		session := config.SSHSession{
			AuthMethod: "key",
			KeyPath:    "nonexistent_key",
		}
		// This should fail because the key file doesn't exist
		_, err := m.getAuth(session)
		if err == nil {
			t.Error("expected error for nonexistent key file")
		}
	})
}

func TestGetHostKeyCallback(t *testing.T) {
	m := &SSHManager{}

	// Setup a temporary base directory for config
	tmpDir, _ := os.MkdirTemp("", "ssh_test_config")
	defer os.RemoveAll(tmpDir)

	// We can't easily mock config.GetBaseDir() without changing its implementation,
	// but it uses OSTENIA_HOME env var.
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	callback := m.getHostKeyCallback()
	if callback == nil {
		t.Fatal("expected non-nil callback")
	}

	knownHostsPath := filepath.Join(tmpDir, "known_hosts")
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		t.Error("expected known_hosts file to be created")
	}
}
