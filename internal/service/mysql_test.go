package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"testing"
)

func TestMySQLConfig(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("UpdateMySQLConfig", func(t *testing.T) {
		err := UpdateMySQLConfig(tempDir, filepath.Join(tempDir, "data"), filepath.Join(tempDir, "tmp"), 3306)
		if err != nil {
			t.Errorf("UpdateMySQLConfig failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(tempDir, "my.ini")); err != nil {
			t.Error("my.ini not created")
		}
	})

	t.Run("InitializeMySQLDataDir", func(t *testing.T) {
		origExecutor := utils.Executor
		utils.Executor = &mockExecutor{output: []byte("")}
		defer func() { utils.Executor = origExecutor }()

		binDir := filepath.Join(tempDir, "bin")
		os.MkdirAll(binDir, 0755)

		// Create dummy empty mysqld.exe so Stat succeeds
		os.WriteFile(filepath.Join(binDir, "mysqld.exe"), []byte(""), 0755)

		err := InitializeMySQLDataDir(binDir, tempDir, filepath.Join(tempDir, "data"), filepath.Join(tempDir, "my.ini"))
		if err != nil {
			t.Errorf("InitializeMySQLDataDir failed: %v", err)
		}
	})
}

