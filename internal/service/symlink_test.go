package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"testing"
)

func TestSymlinkManager(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &mockExecutor{}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin", "php")
	os.MkdirAll(binDir, 0755)

	versionDir := filepath.Join(binDir, "php-8.2.0")
	os.MkdirAll(versionDir, 0755)

	s := NewSymlinkManager()

	t.Run("SwitchVersion", func(t *testing.T) {
		// Set Ostenia home for config
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		err := s.SwitchVersion("php", "php-8.2.0")
		if err != nil {
			t.Errorf("SwitchVersion failed: %v", err)
		}
	})

    t.Run("SwitchVersion_NonExistent", func(t *testing.T) {
        err := s.SwitchVersion("php", "non-existent")
        if err == nil {
            t.Error("Expected error for non-existent directory")
        }
    })
}
