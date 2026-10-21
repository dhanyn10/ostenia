package service

import (
	"errors"
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"strings"
	"syscall"
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

func TestValidateMySQLDataDir(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Non-existent dir returns nil
	if err := ValidateMySQLDataDir(filepath.Join(tempDir, "nonexistent")); err != nil {
		t.Errorf("Expected nil for nonexistent dir, got: %v", err)
	}

	// 2. Existing writable dir without ibdata1 returns nil
	dataDir := filepath.Join(tempDir, "data")
	_ = os.MkdirAll(dataDir, 0755)
	if err := ValidateMySQLDataDir(dataDir); err != nil {
		t.Errorf("Expected nil for clean writable dir, got: %v", err)
	}

	// 3. Existing writable dir with writable ibdata1 returns nil
	ibdataPath := filepath.Join(dataDir, "ibdata1")
	_ = os.WriteFile(ibdataPath, []byte("data"), 0644)
	if err := ValidateMySQLDataDir(dataDir); err != nil {
		t.Errorf("Expected nil for dir with writable ibdata1, got: %v", err)
	}
}

func TestDiagnoseFileLockError(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "ibdata1")
	_ = os.WriteFile(filePath, []byte("test"), 0600)

	// 1. Sharing violation (errno 32)
	errSharing := &os.PathError{Op: "open", Path: filePath, Err: syscall.Errno(32)}
	diagSharing := diagnoseFileLockError("ibdata1", errSharing, filePath)
	if !strings.Contains(diagSharing.Error(), "currently locked by another process") {
		t.Errorf("Expected sharing violation message, got: %v", diagSharing)
	}

	// 2. Lock violation (errno 33)
	errLock := &os.PathError{Op: "open", Path: filePath, Err: syscall.Errno(33)}
	diagLock := diagnoseFileLockError("ibdata1", errLock, filePath)
	if !strings.Contains(diagLock.Error(), "currently locked by another process") {
		t.Errorf("Expected lock violation message, got: %v", diagLock)
	}

	// 3. Access denied (errno 5) - writable file succeeds after chmod
	errDenied := &os.PathError{Op: "open", Path: filePath, Err: syscall.Errno(5)}
	diagDenied := diagnoseFileLockError("ibdata1", errDenied, filePath)
	if diagDenied != nil {
		t.Errorf("Expected nil when retry succeeds on writable file, got: %v", diagDenied)
	}

	// 4. Access denied on non-existent file
	errDenied2 := &os.PathError{Op: "open", Path: filepath.Join(tempDir, "missing"), Err: syscall.Errno(5)}
	diagDenied2 := diagnoseFileLockError("missing", errDenied2, filepath.Join(tempDir, "missing"))
	if !strings.Contains(diagDenied2.Error(), "Permission Denied") {
		t.Errorf("Expected permission denied message, got: %v", diagDenied2)
	}

	// 5. Generic error
	genericErr := errors.New("disk failure")
	diagGeneric := diagnoseFileLockError("ibdata1", genericErr, filePath)
	if !strings.Contains(diagGeneric.Error(), "failed to access database file") {
		t.Errorf("Expected generic error message, got: %v", diagGeneric)
	}
}
