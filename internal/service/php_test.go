package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPHPConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Setup dummy php.ini-development
	os.WriteFile(filepath.Join(tempDir, "php.ini-development"), []byte("memory_limit = 128M\n;extension=curl"), 0644)

	t.Run("UpdatePHPConfig", func(t *testing.T) {
		err := UpdatePHPConfig(tempDir)
		if err != nil {
			t.Errorf("UpdatePHPConfig failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(tempDir, "php.ini")); err != nil {
			t.Error("php.ini not created")
		}
	})

	t.Run("Extensions", func(t *testing.T) {
		// Mock ext dir
		extDir := filepath.Join(tempDir, "ext")
		os.MkdirAll(extDir, 0755)
		os.WriteFile(filepath.Join(extDir, "php_curl.dll"), []byte(""), 0644)

		exts, _ := GetPHPExtensions(tempDir)
		found := false
		for _, e := range exts {
			if e.Name == "curl" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected curl extension to be found")
		}

		_ = TogglePHPExtension(tempDir, "curl", true)
	})
}
