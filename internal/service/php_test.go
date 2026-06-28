package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPHPConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "php_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	iniPath := filepath.Join(tmpDir, "php.ini")
	initialContent := ";extension=openssl\n;extension=curl\n"
	os.WriteFile(iniPath, []byte(initialContent), 0644)

	// Test UpdatePHPConfig
	err = UpdatePHPConfig(tmpDir)
	if err != nil {
		t.Fatalf("UpdatePHPConfig failed: %v", err)
	}

	content, _ := os.ReadFile(iniPath)
	if !strings.Contains(string(content), "extension=openssl") || strings.HasPrefix(strings.TrimSpace(string(content)), ";extension=openssl") {
		// openssl should be enabled (semicolon removed)
		if strings.Contains(string(content), ";extension=openssl") {
			t.Error("Expected openssl to be enabled")
		}
	}

	// Test GetPHPExtensions
	exts, err := GetPHPExtensions(tmpDir)
	if err != nil {
		t.Fatalf("GetPHPExtensions failed: %v", err)
	}
	foundOpenssl := false
	for _, e := range exts {
		if e.Name == "openssl" {
			foundOpenssl = true
			if !e.Enabled {
				t.Error("Expected openssl to be enabled")
			}
		}
	}
	if !foundOpenssl {
		t.Error("openssl extension not found in list")
	}

	// Test TogglePHPExtension
	err = TogglePHPExtension(tmpDir, "openssl", false)
	if err != nil {
		t.Fatalf("TogglePHPExtension failed: %v", err)
	}
	exts, _ = GetPHPExtensions(tmpDir)
	for _, e := range exts {
		if e.Name == "openssl" && e.Enabled {
			t.Error("Expected openssl to be disabled")
		}
	}
}
