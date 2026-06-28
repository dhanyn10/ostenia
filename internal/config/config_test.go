package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// Setup a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Since LoadConfig/SaveConfig uses os.Executable() path, we can't easily override it
	// without refactoring. But we can test GetBaseDir with environment variables.

	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	baseDir := GetBaseDir()
	if baseDir != tmpDir {
		t.Errorf("Expected baseDir to be %s, got %s", tmpDir, baseDir)
	}

	cfg := &Config{
		BaseDir: tmpDir,
		WWWRoot: filepath.Join(tmpDir, "www"),
		Proxies: make(map[string]int),
	}
	cfg.Proxies["test"] = 8080

	// We can't easily test SaveConfig/LoadConfig here because they depend on os.Executable()
}
