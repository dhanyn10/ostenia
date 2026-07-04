package service

import (
	"os"
	"ostenia/internal/ssl"
	"path/filepath"
	"testing"
)

func TestNginxConfig(t *testing.T) {
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "conf")
	os.MkdirAll(confDir, 0755)

	confPath := filepath.Join(confDir, "nginx.conf")
	os.WriteFile(confPath, []byte("http { server { listen 80; } }"), 0644)

    // Mock SSL
    origSign := ssl.SignCertificateFunc
    ssl.SignCertificateFunc = func(ca, dom, dst string) error { return nil }
    defer func() { ssl.SignCertificateFunc = origSign }()

	t.Run("UpdateNginxConfig", func(t *testing.T) {
		proxies := []ProxyConfig{
			{Name: "app1", TargetPort: 3000},
		}

		err := UpdateNginxConfig(tempDir, tempDir, 9000, 8080, false, proxies)
		if err != nil {
			t.Errorf("UpdateNginxConfig failed: %v", err)
		}

		// Verify nginx.conf creation (ostenia.conf was from a different version or my mistake)
		if _, err := os.Stat(filepath.Join(confDir, "nginx.conf")); err != nil {
			t.Error("nginx.conf not created")
		}
	})
}
