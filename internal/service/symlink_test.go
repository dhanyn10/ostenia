package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSymlinkManager(t *testing.T) {
	mockSys := NewMockSystem()
	sm := NewSymlinkManager(mockSys)

	t.Run("SwitchVersion", func(t *testing.T) {
		tempDir := t.TempDir()
		binDir := filepath.Join(tempDir, "bin", "php")
		os.MkdirAll(binDir, 0755)

		targetDir := filepath.Join(binDir, "php-8.2.0")
		os.MkdirAll(targetDir, 0755)

		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		err := sm.SwitchVersion("php", "php-8.2.0")
		if err != nil {
			t.Errorf("SwitchVersion failed: %v", err)
		}
	})
}
