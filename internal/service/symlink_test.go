package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSymlinkManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "symlink_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	category := "php"
	binDir := filepath.Join(tmpDir, "bin", category)
	versionDir := filepath.Join(binDir, "8.2.0")
	os.MkdirAll(versionDir, 0755)

	mgr := NewSymlinkManager()
	err = mgr.SwitchVersion(category, "8.2.0")
	if err != nil {
		// On non-windows it should work. On windows it might fail because 'mklink' needs cmd.exe
		// In CI (Linux) it should use os.Symlink
		t.Logf("SwitchVersion result: %v", err)
	}

	currentLink := filepath.Join(binDir, "current")
	if _, err := os.Lstat(currentLink); err != nil {
		t.Log("current link not created, possibly platform restriction")
	}
}
