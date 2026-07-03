package service

import (
	"os"
	"path/filepath"
	"testing"
    "ostenia/internal/ssl"
)

func TestApacheConfig(t *testing.T) {
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "conf")
	extraDir := filepath.Join(confDir, "extra")
	os.MkdirAll(extraDir, 0755)

	confPath := filepath.Join(confDir, "httpd.conf")
	os.WriteFile(confPath, []byte("Listen 80\n#LoadModule rewrite_module"), 0644)

    origRSA := ssl.RSAKeySize
    ssl.RSAKeySize = 1024
    defer func() { ssl.RSAKeySize = origRSA }()

	t.Run("GenerateVHost", func(t *testing.T) {
		res := GenerateVHost("test", "/path", 80)
		if res == "" {
			t.Error("Expected non-empty VHost")
		}
	})

	t.Run("GenerateProxyVHost", func(t *testing.T) {
		res := GenerateProxyVHost("test", 3000, 80, true, tempDir)
		if res == "" {
			t.Error("Expected non-empty Proxy VHost")
		}
	})

	t.Run("UpdateApacheConfig", func(t *testing.T) {
		err := UpdateApacheConfig(tempDir, "", "", "VHOST_CONTENT", 8080, tempDir, 9000, true)
		if err != nil {
			t.Errorf("UpdateApacheConfig failed: %v", err)
		}

		// Verify file creation
		if _, err := os.Stat(filepath.Join(extraDir, "httpd-ostenia-php.conf")); err != nil {
			t.Error("PHP conf not created")
		}
		if _, err := os.Stat(filepath.Join(extraDir, "httpd-ostenia-ssl.conf")); err != nil {
			t.Error("SSL conf not created")
		}
	})
}
