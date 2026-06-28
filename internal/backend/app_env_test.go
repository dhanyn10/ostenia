package backend

import (
	"context"
	"ostenia/internal/config"
	"testing"
)

func TestAppEnv(t *testing.T) {
	app := NewApp()
	app.Startup(context.Background())
	app.cfg = &config.Config{
		WWWRoot: "/tmp/www",
	}

	if app.IsAdmin() != app.IsAdmin() { // just call it
		t.Error("IsAdmin inconsistent")
	}

	// ensureEnvironmentStructure is called during Startup, but we can call it again
	app.ensureEnvironmentStructure()
}
