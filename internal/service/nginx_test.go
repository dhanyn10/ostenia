package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNginxConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nginx_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nginxPath := filepath.Join(tmpDir, "nginx")
	wwwRoot := filepath.Join(tmpDir, "www")
	os.MkdirAll(filepath.Join(nginxPath, "conf"), 0755)

	// Mock config base dir to satisfy UpdateNginxConfig's use of config.GetBaseDir()
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	proxies := []ProxyConfig{
		{Name: "app", TargetPort: 3000},
	}

	err = UpdateNginxConfig(nginxPath, wwwRoot, 9000, 80, false, proxies)
	if err != nil {
		t.Fatalf("UpdateNginxConfig failed: %v", err)
	}

	confPath := filepath.Join(nginxPath, "conf", "nginx.conf")
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		t.Fatal("nginx.conf was not created")
	}

	content, _ := os.ReadFile(confPath)
	if !strings.Contains(string(content), "listen       80;") {
		t.Error("Expected listen 80 in nginx.conf")
	}
	if !strings.Contains(string(content), "server_name  app.test") {
		t.Error("Expected proxy server_name in nginx.conf")
	}
}
