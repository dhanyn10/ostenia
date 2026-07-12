package service

import (
	"os"
	"path/filepath"
	"strings"
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

	t.Run("UpdateApacheConfig_Full", func(t *testing.T) {
		apachePath := filepath.Join(tempDir, "apache_full")
		os.MkdirAll(filepath.Join(apachePath, "conf", "extra"), 0755)
		os.WriteFile(filepath.Join(apachePath, "conf", "httpd.conf"), []byte("#LoadModule rewrite_module modules/mod_rewrite.so\nListen 80"), 0644)

		err := UpdateApacheConfig(apachePath, "php82.dll", "php", "VHost content", 8080, filepath.Join(tempDir, "www"), 9000, true)
		if err != nil {
			t.Errorf("UpdateApacheConfig failed: %v", err)
		}

		conf, _ := os.ReadFile(filepath.Join(apachePath, "conf", "httpd.conf"))
		if !strings.Contains(string(conf), "LoadModule rewrite_module") {
			t.Error("Expected rewrite_module to be enabled")
		}
		if !strings.Contains(string(conf), "Listen 8080") {
			t.Error("Expected port 8080")
		}
	})

	t.Run("UpdateApacheConfig_Minimal", func(t *testing.T) {
		apachePath := filepath.Join(tempDir, "apache_min")
		os.MkdirAll(filepath.Join(apachePath, "conf", "extra"), 0755)
		os.WriteFile(filepath.Join(apachePath, "conf", "httpd.conf"), []byte(""), 0644)

		err := UpdateApacheConfig(apachePath, "", "", "", 0, "/www", 0, false)
		if err != nil {
			t.Errorf("UpdateApacheConfig failed: %v", err)
		}
	})

	t.Run("UpdateApacheConfig_WithVHosts", func(t *testing.T) {
		apachePath := filepath.Join(tempDir, "apache_vhosts")
		os.MkdirAll(filepath.Join(apachePath, "conf", "extra"), 0755)
		os.WriteFile(filepath.Join(apachePath, "conf", "httpd.conf"), []byte(""), 0644)

		err := UpdateApacheConfig(apachePath, "", "", "VHosts", 80, "/www", 0, false)
		if err != nil {
			t.Errorf("UpdateApacheConfig failed: %v", err)
		}

		conf, _ := os.ReadFile(filepath.Join(apachePath, "conf", "httpd.conf"))
		if !strings.Contains(string(conf), "Include conf/extra/httpd-ostenia-vhosts.conf") {
			t.Error("Expected vhosts include")
		}
	})

	t.Run("GenerateVHost_DefaultPort", func(t *testing.T) {
		vhost := GenerateVHost("example", "/path", 0)
		if !strings.Contains(vhost, "<VirtualHost *:80>") {
			t.Error("Expected default port 80")
		}
	})

	t.Run("UpdateApacheConfig_Error", func(t *testing.T) {
		err := UpdateApacheConfig("/invalid/path", "", "", "", 0, "/www", 0, false)
		if err == nil {
			t.Error("Expected error for invalid path")
		}
	})
}
