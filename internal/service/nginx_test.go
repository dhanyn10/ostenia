package service

import (
	"os"
	"ostenia/internal/ssl"
	"path/filepath"
	"strings"
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
	ssl.SignCertificateFunc = func(ca, dom, dst string) error {
		// Mock cert files
		os.WriteFile(filepath.Join(dst, "localhost.crt"), []byte("cert"), 0644)
		os.WriteFile(filepath.Join(dst, "localhost.key"), []byte("key"), 0644)
		return nil
	}
	defer func() { ssl.SignCertificateFunc = origSign }()

	t.Run("UpdateNginxConfig_HTTP", func(t *testing.T) {
		proxies := []ProxyConfig{
			{Name: "app1", TargetPort: 3000},
		}

		err := UpdateNginxConfig(tempDir, tempDir, 9000, 80, false, proxies)
		if err != nil {
			t.Errorf("UpdateNginxConfig failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(confDir, "nginx.conf")); err != nil {
			t.Error("nginx.conf not created")
		}
	})

	t.Run("UpdateNginxConfig_HTTPS", func(t *testing.T) {
		err := UpdateNginxConfig(tempDir, tempDir, 9000, 443, true, nil)
		if err != nil {
			t.Errorf("UpdateNginxConfig failed: %v", err)
		}

		content, _ := os.ReadFile(filepath.Join(confDir, "nginx.conf"))
		if !strings.Contains(string(content), "ssl_certificate") {
			t.Error("Expected SSL configuration in nginx.conf")
		}
	})

	t.Run("UpdateNginxConfig_DefaultPort", func(t *testing.T) {
		err := UpdateNginxConfig(tempDir, tempDir, 9000, 0, false, nil)
		if err != nil {
			t.Errorf("UpdateNginxConfig failed: %v", err)
		}
	})

	t.Run("UpdateNginxConfig_HTTPS_CertNotFound", func(t *testing.T) {
		// Mock SignCertificate to NOT create cert files
		origSign := ssl.SignCertificateFunc
		ssl.SignCertificateFunc = func(ca, dom, dst string) error {
			return nil
		}
		defer func() { ssl.SignCertificateFunc = origSign }()

		// Remove any existing localhost.crt from tempDir/ssl to trigger fallback
		_ = os.Remove(filepath.Join(tempDir, "ssl", "localhost.crt"))

		err := UpdateNginxConfig(tempDir, tempDir, 9000, 443, true, nil)
		if err != nil {
			t.Errorf("UpdateNginxConfig failed on CertNotFound: %v", err)
		}
	})
}
