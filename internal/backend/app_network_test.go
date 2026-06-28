package backend

import (
	"context"
	"os"
	"ostenia/internal/config"
	"path/filepath"
	"testing"
)

func TestAppNetwork(t *testing.T) {
	app := NewApp()
	app.Startup(context.Background())

	tmpDir, _ := os.MkdirTemp("", "network_test")
	defer os.RemoveAll(tmpDir)

	wwwRoot := filepath.Join(tmpDir, "www")
	os.MkdirAll(wwwRoot, 0755)
	os.MkdirAll(filepath.Join(wwwRoot, "app1"), 0755)

	app.cfg = &config.Config{
		WWWRoot: wwwRoot,
		Proxies: map[string]int{
			"app1": 3000,
		},
	}

	apps := app.GetProxyApps()
	if len(apps) == 0 {
		t.Error("Expected at least 1 proxy app")
	}

	statuses := app.CheckProxyPorts()
	if len(statuses) == 0 {
		t.Error("Expected proxy statuses")
	}
}
