package config

import (
	"os"
	"testing"
)

func TestSSHSessions(t *testing.T) {
	// Setup temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ostenia-ssh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock OSTENIA_HOME to control GetBaseDir which is used by getSSHSessionsPath
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	t.Run("Empty Sessions", func(t *testing.T) {
		sessions, err := LoadSSHSessions()
		if err != nil {
			t.Fatalf("LoadSSHSessions() error = %v", err)
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %v", len(sessions))
		}
	})

	t.Run("Add and Load Sessions", func(t *testing.T) {
		session := SSHSession{
			ID:       "test-id",
			Name:     "Test Session",
			Password: "secret-password",
		}

		err := AddSSHSession(session)
		if err != nil {
			t.Fatalf("AddSSHSession() error = %v", err)
		}

		sessions, err := LoadSSHSessions()
		if err != nil {
			t.Fatalf("LoadSSHSessions() error = %v", err)
		}

		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %v", len(sessions))
		}

		if sessions[0].ID != "test-id" {
			t.Errorf("Expected ID test-id, got %v", sessions[0].ID)
		}

		if sessions[0].Password != "secret-password" {
			t.Errorf("Expected password secret-password, got %v", sessions[0].Password)
		}
	})

	t.Run("Update Session", func(t *testing.T) {
		session := SSHSession{
			ID:   "test-id",
			Name: "Updated Session",
		}

		err := UpdateSSHSession(session)
		if err != nil {
			t.Fatalf("UpdateSSHSession() error = %v", err)
		}

		sessions, err := LoadSSHSessions()
		if err != nil {
			t.Fatalf("LoadSSHSessions() error = %v", err)
		}

		if sessions[0].Name != "Updated Session" {
			t.Errorf("Expected name Updated Session, got %v", sessions[0].Name)
		}
	})

	t.Run("Delete Session", func(t *testing.T) {
		err := DeleteSSHSession("test-id")
		if err != nil {
			t.Fatalf("DeleteSSHSession() error = %v", err)
		}

		sessions, err := LoadSSHSessions()
		if err != nil {
			t.Fatalf("LoadSSHSessions() error = %v", err)
		}

		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %v", len(sessions))
		}
	})
}
