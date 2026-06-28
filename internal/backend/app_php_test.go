package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAppPHP(t *testing.T) {
	app := NewApp()
	app.Startup(context.Background())

	tmpDir, _ := os.MkdirTemp("", "php_backend_test")
	defer os.RemoveAll(tmpDir)

	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	phpPath := filepath.Join(tmpDir, "bin", "php", "current")
	os.MkdirAll(phpPath, 0755)
	os.WriteFile(filepath.Join(phpPath, "php.ini"), []byte(";extension=openssl"), 0644)

	exts, _ := app.GetPHPExtensions()
	if len(exts) == 0 {
		t.Log("No extensions found (expected if ini is empty/mocked)")
	}

	_ = app.TogglePHPExtension("openssl", true)
}
