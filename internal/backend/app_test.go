package backend

import (
	"context"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("Expected App instance, got nil")
	}
}

func TestAppStartup(t *testing.T) {
	// Setup dummy frontend/dist because Startup might indirectly trigger things that need it
	// though Startup itself seems to just init objects.
	app := NewApp()
	ctx := context.Background()

	// Startup calls LoadConfig and ensureEnvironmentStructure
	// We might need to mock config if we want to be safe
	app.Startup(ctx)

	if app.ctx != ctx {
		t.Errorf("Expected context to be set")
	}
	if app.downloader == nil {
		t.Error("Expected downloader to be initialized")
	}
	if app.orchestrator == nil {
		t.Error("Expected orchestrator to be initialized")
	}
}
