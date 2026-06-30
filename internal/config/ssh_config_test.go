package config

import (
	"os"
	"testing"
)

func setupSSHTest(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "ostenia-ssh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	os.Setenv("OSTENIA_HOME", tmpDir)
	return tmpDir
}

func TestSSHSessions(t *testing.T) {
	tmpDir := setupSSHTest(t)
	defer os.RemoveAll(tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	t.Run("Empty Sessions", testEmptySessions)
	t.Run("Add and Load Sessions", testAddLoadSessions)
	t.Run("Update Session", testUpdateSession)
	t.Run("Delete Session", testDeleteSession)
}

func testEmptySessions(t *testing.T) {
	sessions, err := LoadSSHSessions()
	if err != nil {
		t.Fatalf("LoadSSHSessions() error = %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %v", len(sessions))
	}
}

func testAddLoadSessions(t *testing.T) {
	session := SSHSession{
		ID:       "test-id",
		Name:     "Test Session",
		Password: "secret-password",
	}

	if err := AddSSHSession(session); err != nil {
		t.Fatalf("AddSSHSession() error = %v", err)
	}

	sessions, err := LoadSSHSessions()
	if err != nil {
		t.Fatalf("LoadSSHSessions() error = %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %v", len(sessions))
	}

	if sessions[0].ID != "test-id" {
		t.Errorf("Expected ID test-id, got %v", sessions[0].ID)
	}

	if sessions[0].Password != "secret-password" {
		t.Errorf("Expected password secret-password, got %v", sessions[0].Password)
	}
}

func testUpdateSession(t *testing.T) {
	session := SSHSession{
		ID:   "test-id",
		Name: "Updated Session",
	}

	if err := UpdateSSHSession(session); err != nil {
		t.Fatalf("UpdateSSHSession() error = %v", err)
	}

	sessions, err := LoadSSHSessions()
	if err != nil {
		t.Fatalf("LoadSSHSessions() error = %v", err)
	}

	if len(sessions) > 0 && sessions[0].Name != "Updated Session" {
		t.Errorf("Expected name Updated Session, got %v", sessions[0].Name)
	}
}

func testDeleteSession(t *testing.T) {
	if err := DeleteSSHSession("test-id"); err != nil {
		t.Fatalf("DeleteSSHSession() error = %v", err)
	}

	sessions, err := LoadSSHSessions()
	if err != nil {
		t.Fatalf("LoadSSHSessions() error = %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %v", len(sessions))
	}
}
