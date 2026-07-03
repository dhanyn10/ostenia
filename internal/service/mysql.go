package service

import (
	"fmt"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"strings"
)

// UpdateMySQLConfig generates or updates the my.ini file for MySQL.
func UpdateMySQLConfig(mysqlBaseDir string, dataDir string, tmpDir string, port int) error {
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
func InitializeMySQLDataDir(mysqlBinDir string, mysqlBaseDir string, dataDir string, iniPath string) error {
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
