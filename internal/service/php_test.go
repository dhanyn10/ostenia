package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
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

	t.Run("GetPHPVersion", func(t *testing.T) {
		origExecutor := utils.Executor
		utils.Executor = &testutil.MockExecutor{Output: "8.2.0"}
		defer func() { utils.Executor = origExecutor }()

		php82 := filepath.Join(tempDir, "php-8.2.0")
		os.MkdirAll(php82, 0755)
		os.WriteFile(filepath.Join(php82, "php.exe"), []byte(""), 0755)

		v, _ := GetPHPVersion(php82)
		if v != "8.2.0" {
			t.Errorf("Expected 8.2.0, got %s", v)
		}
	})

	t.Run("isPHPExtensionEnabled", func(t *testing.T) {
		lines := []string{"extension=curl", ";extension=openssl"}
		if !isPHPExtensionEnabled(lines, "curl") {
			t.Error("Expected curl to be enabled")
		}
		if isPHPExtensionEnabled(lines, "openssl") {
			t.Error("Expected openssl to be disabled")
		}
	})
}
