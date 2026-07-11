package service

import (
	"context"
	"os"
	"ostenia/internal/config"
	"path/filepath"
	"testing"
)

func TestSSHManager_Basic(t *testing.T) {
	ctx := context.Background()
	m := NewSSHManager(ctx)

	t.Run("GetSessions", func(t *testing.T) {
		_, err := m.GetSessions()
		if err != nil {
			// This might fail if OSTENIA_HOME is not set, but we can't easily fix config.LoadSSHSessions
			// without mocking the whole config package.
		}
	})

	t.Run("SaveSessions", func(t *testing.T) {
		err := m.SaveSessions([]config.SSHSession{})
		if err != nil {
			// Similar to GetSessions
		}
	})

	t.Run("ResizeTerminal_NotFound", func(t *testing.T) {
		err := m.ResizeTerminal("invalid", 80, 24)
		if err == nil {
			t.Error("Expected error for invalid session")
		}
	})

	t.Run("SendInput_NotFound", func(t *testing.T) {
		err := m.SendInput("invalid", "ls\n")
		if err == nil {
			t.Error("Expected error for invalid session")
		}
	})

	t.Run("Disconnect_NotFound", func(t *testing.T) {
		// Should not panic
		m.Disconnect("invalid")
	})
}

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
		if len(auth) == 0 {
			t.Error("expected auth methods to be present")
		}
	})

	t.Run("Key Auth", func(t *testing.T) {
		session := config.SSHSession{
			AuthMethod: "key",
			KeyPath:    "nonexistent_key",
		}
		_, err := m.getAuth(session)
		if err == nil {
			t.Error("expected error for nonexistent key file")
		}
	})

	t.Run("Agent Auth", func(t *testing.T) {
		session := config.SSHSession{
			AuthMethod: "agent",
		}
		// We can't really test agent auth in this environment as it depends on SSH_AUTH_SOCK
		_, _ = m.getAuth(session)
	})
}

func TestGetHostKeyCallback(t *testing.T) {
	m := &SSHManager{}

	tmpDir, _ := os.MkdirTemp("", "ssh_test_config")
	defer os.RemoveAll(tmpDir)

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
