package service

import (
	"fmt"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"strings"
	"syscall"
)

// UpdateMySQLConfig generates or updates the my.ini file for MySQL.
func UpdateMySQLConfig(mysqlBaseDir, dataDir, tmpDir string, port int) error {
	iniPath := filepath.Join(mysqlBaseDir, "my.ini")

	// Ensure data and tmp directories exist
	_ = os.MkdirAll(dataDir, 0755)
	_ = os.MkdirAll(tmpDir, 0755)

	content := fmt.Sprintf(`
[mysqld]
# Path to installation directory. All paths are usually relative to this.
basedir="%s"
# Path to the database root
datadir="%s"
# Port the MySQL server will listen on
port=%d
# Path to the temporary files directory
tmpdir="%s"
# Enable the InnoDB storage engine
default_storage_engine=InnoDB
# Allow connections from any host (for development)
bind-address=0.0.0.0
# Character set settings
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci

# Error logging
log-error="%s"
# PID file
pid-file="%s"

[mysql]
default-character-set=utf8mb4

[client]
port=%d
default-character-set=utf8mb4
`,
		strings.ReplaceAll(mysqlBaseDir, "\\", "/"),
		strings.ReplaceAll(dataDir, "\\", "/"),
		port,
		strings.ReplaceAll(tmpDir, "\\", "/"),
		strings.ReplaceAll(filepath.Join(mysqlBaseDir, "mysql_error.log"), "\\", "/"),
		strings.ReplaceAll(filepath.Join(mysqlBaseDir, "mysqld.pid"), "\\", "/"),
		port,
	)

	return os.WriteFile(iniPath, []byte(content), 0644)
}

// InitializeMySQLDataDir runs mysql_install_db (or equivalent) if the data directory is empty.
func InitializeMySQLDataDir(mysqlBinDir, mysqlBaseDir, dataDir, iniPath string) error {
	// Check if data directory is empty
	entries, err := os.ReadDir(dataDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading MySQL data directory: %w", err)
	}
	if len(entries) > 0 {
		fmt.Println("[MySQL] Data directory is not empty, skipping initialization.")
		return nil // Data directory already initialized
	}

	fmt.Println("[MySQL] Data directory is empty, initializing MySQL data directory...")

	mysqlInstallDbPath := filepath.Join(mysqlBinDir, "mysql_install_db.exe") // For older MySQL versions
	mysqldPath := filepath.Join(mysqlBinDir, "mysqld.exe")                   // For newer MySQL versions (8.0+)

	var cmd *exec.Cmd

	// Try mysql_install_db first (older versions)
	if _, err := os.Stat(mysqlInstallDbPath); err == nil {
		cmd = utils.Executor.Command(mysqlInstallDbPath, "--defaults-file="+iniPath, "--basedir="+mysqlBaseDir, "--datadir="+dataDir, "--console")
	} else if _, err := os.Stat(mysqldPath); err == nil {
		// For MySQL 8.0+, use mysqld --initialize-insecure
		cmd = utils.Executor.Command(mysqldPath, "--defaults-file="+iniPath, "--initialize-insecure", "--console")
	} else {
		return fmt.Errorf("neither mysql_install_db.exe nor mysqld.exe found for initialization")
	}

	if cmd == nil {
		return fmt.Errorf("failed to prepare MySQL initialization command")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize MySQL data directory: %w", err)
	}

	fmt.Println("[MySQL] MySQL data directory initialized successfully.")
	return nil
}

// ValidateMySQLDataDir checks whether the MySQL data directory and critical files (such as ibdata1)
// can be accessed and written to, distinguishing between sharing violations (locked by another process)
// and actual permission denied / UAC issues.
func ValidateMySQLDataDir(dataDir string) error {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return nil // Not created yet, will be initialized cleanly
	}

	// 1. Check directory writability
	testFile := filepath.Join(dataDir, ".ostenia_perm_check")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("MySQL data directory '%s' is not writable (Permission Denied). Administrator privileges (UAC) may be required: %w", dataDir, err)
	}
	_ = os.Remove(testFile)

	// 2. Check ibdata1 if it exists
	ibdataPath := filepath.Join(dataDir, "ibdata1")
	if _, err := os.Stat(ibdataPath); err == nil {
		f, err := os.OpenFile(ibdataPath, os.O_RDWR, 0666)
		if err != nil {
			return diagnoseFileLockError("ibdata1", err, ibdataPath)
		}
		_ = f.Close()
	}

	return nil
}

func diagnoseFileLockError(fileName string, err error, fullPath string) error {
	if pathErr, ok := err.(*os.PathError); ok {
		if errno, ok := pathErr.Err.(syscall.Errno); ok {
			const (
				errorAccessDenied     syscall.Errno = 5
				errorSharingViolation syscall.Errno = 32
				errorLockViolation    syscall.Errno = 33
			)
			switch errno {
			case errorSharingViolation, errorLockViolation:
				return fmt.Errorf("database file '%s' is currently locked by another process (Sharing Violation). Please ensure no other MySQL instances are running", fileName)
			case errorAccessDenied:
				// Attempt to clear Read-Only attribute if set and retry
				_ = os.Chmod(fullPath, 0666)
				if retryF, retryErr := os.OpenFile(fullPath, os.O_RDWR, 0666); retryErr == nil {
					_ = retryF.Close()
					return nil
				}
				return fmt.Errorf("access denied opening database file '%s' (Permission Denied). Administrator privileges (UAC) may be required", fileName)
			}
		}
	}
	return fmt.Errorf("failed to access database file '%s': %w", fileName, err)
}
