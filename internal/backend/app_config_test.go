package backend

import (
	"context"
	"ostenia/internal/config"
	"testing"
)

func TestAppConfig(t *testing.T) {
	app := NewApp()
	app.Startup(context.Background())

	app.cfg = &config.Config{
		DefaultEditor: "notepad.exe",
	}

	cfg := app.GetConfig()
	if cfg.DefaultEditor != "notepad.exe" {
		t.Errorf("Expected notepad.exe, got %s", cfg.DefaultEditor)
	}

	app.UpdateActiveTab("plugins")
}
