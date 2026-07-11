package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApacheConfig(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("GenerateVHost", func(t *testing.T) {
		vhost := GenerateVHost("example", "/var/www/html", 80)
		if vhost == "" {
			t.Error("Expected non-empty VHost config")
		}
	})

	t.Run("GenerateProxyVHost", func(t *testing.T) {
		vhost := GenerateProxyVHost("example", 3000, 80, true, "/ssl")
		if vhost == "" {
			t.Error("Expected non-empty Proxy VHost config")
		}
	})

	t.Run("UpdateApacheConfig", func(t *testing.T) {
		apachePath := filepath.Join(tempDir, "apache")
		os.MkdirAll(filepath.Join(apachePath, "conf", "extra"), 0755)

		// Create dummy original config
		os.WriteFile(filepath.Join(apachePath, "conf", "httpd.conf"), []byte("Listen 80"), 0644)

		err := UpdateApacheConfig(apachePath, "php82.dll", "php", "VHost content", 80, filepath.Join(tempDir, "www"), 9000, true)
		if err != nil {
			t.Errorf("UpdateApacheConfig failed: %v", err)
		}
	})
}
