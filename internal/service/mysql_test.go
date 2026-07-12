package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
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
		utils.Executor = &testutil.MockExecutor{Output: ""}
		defer func() { utils.Executor = origExecutor }()

		binDir := filepath.Join(tempDir, "bin")
		os.MkdirAll(binDir, 0755)

		os.WriteFile(filepath.Join(binDir, "mysqld.exe"), []byte(""), 0755)

		err := InitializeMySQLDataDir(binDir, tempDir, filepath.Join(tempDir, "data"), filepath.Join(tempDir, "my.ini"))
		if err != nil {
			t.Errorf("InitializeMySQLDataDir failed: %v", err)
		}

		// Test skip when not empty
		os.WriteFile(filepath.Join(tempDir, "data", "test"), []byte(""), 0644)
		err = InitializeMySQLDataDir(binDir, tempDir, filepath.Join(tempDir, "data"), filepath.Join(tempDir, "my.ini"))
		if err != nil {
			t.Errorf("InitializeMySQLDataDir skip failed: %v", err)
		}

		// Test older version path
		os.RemoveAll(filepath.Join(tempDir, "data"))
		os.MkdirAll(filepath.Join(tempDir, "data"), 0755)
		os.WriteFile(filepath.Join(binDir, "mysql_install_db.exe"), []byte(""), 0755)
		err = InitializeMySQLDataDir(binDir, tempDir, filepath.Join(tempDir, "data"), filepath.Join(tempDir, "my.ini"))
		if err != nil {
			t.Errorf("InitializeMySQLDataDir (old) failed: %v", err)
		}

		// Test none found
		os.Remove(filepath.Join(binDir, "mysqld.exe"))
		os.Remove(filepath.Join(binDir, "mysql_install_db.exe"))
		err = InitializeMySQLDataDir(binDir, tempDir, filepath.Join(tempDir, "data"), filepath.Join(tempDir, "my.ini"))
		if err == nil {
			t.Error("Expected error when no binaries found")
		}
	})
}
