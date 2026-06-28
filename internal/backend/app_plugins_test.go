package backend

import (
	"context"
	"testing"
)

func TestAppPlugins(t *testing.T) {
	app := NewApp()
	app.Startup(context.Background())

	tasks := app.GetPrerequisites()
	if len(tasks) == 0 {
		t.Error("Expected at least 1 task")
	}

	cat, bin, cur := app.getPluginPaths("PHP")
	if cat != "php" {
		t.Errorf("Expected php, got %s", cat)
	}
	if bin == "" || cur == "" {
		t.Error("Paths should not be empty")
	}
}
