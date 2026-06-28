package config

import (
	"os"
	"testing"
)

func TestSSHSessions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ssh_config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	session := SSHSession{
		ID:   "1",
		Name: "Test",
		User: "user",
		Host: "localhost",
	}

	err = AddSSHSession(session)
	if err != nil {
		t.Fatalf("AddSSHSession failed: %v", err)
	}

	sessions, err := LoadSSHSessions()
	if err != nil {
		t.Fatalf("LoadSSHSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	session.Name = "Updated"
	err = UpdateSSHSession(session)
	if err != nil {
		t.Fatalf("UpdateSSHSession failed: %v", err)
	}

	sessions, _ = LoadSSHSessions()
	if sessions[0].Name != "Updated" {
		t.Errorf("Expected Updated name, got %s", sessions[0].Name)
	}

	err = DeleteSSHSession("1")
	if err != nil {
		t.Fatalf("DeleteSSHSession failed: %v", err)
	}

	sessions, _ = LoadSSHSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
}
