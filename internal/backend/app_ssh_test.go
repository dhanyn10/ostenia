package backend

import (
	"context"
	"ostenia/internal/config"
	"testing"
)

func TestAppSSH(t *testing.T) {
	app := NewApp()
	app.Startup(context.Background())

	// Just test methods that don't need real SSH or Wails runtime
	sessions, _ := app.GetSSHSessions()
	if sessions == nil {
		t.Log("No SSH sessions found")
	}

	session := config.SSHSession{ID: "1", Name: "Test"}
	_ = app.AddSSHSession(session)
	_ = app.UpdateSSHSession(session)
	_ = app.DeleteSSHSession("1")
}
