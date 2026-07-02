package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPlugins_Complete(t *testing.T) {
	ctx := context.Background()
	m := NewManager(ctx)
    m.emit = func(ctx context.Context, eventName string, optionalData ...interface{}) {}

	t.Run("Manager_Basic", func(t *testing.T) {
		m.CancelDownload("test")
	})

	t.Run("DeleteVersion", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		verDir := filepath.Join(tempDir, "bin", "php", "php-8.1.0")
		os.MkdirAll(verDir, 0755)

		_ = m.DeleteVersion("PHP", "8.1.0")
	})

	t.Run("GetInstalledVersionPaths", func(t *testing.T) {
		tempDir := t.TempDir()
		os.MkdirAll(filepath.Join(tempDir, "bin", "php", "php-8.1.0"), 0755)
		os.WriteFile(filepath.Join(tempDir, "bin", "php", "php-8.1.0", "php.exe"), []byte(""), 0644)

		paths := GetInstalledVersionPaths(tempDir, "php", "php.exe")
		if len(paths) == 0 {
			t.Error("Expected versions")
		}
	})

    t.Run("Discovery", func(t *testing.T) {
        _ = GetLatestKnownVersions()
    })
}
