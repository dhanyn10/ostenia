package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// GetInstalledVersionPaths returns a map of version string to its absolute path for a given category.
func GetInstalledVersionPaths(baseDir, category, checkFile string) map[string]string {
	versions := make(map[string]string)
	binDir := filepath.Join(baseDir, "bin", category)
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return versions
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			checkPath := filepath.Join(binDir, entry.Name(), checkFile)

			// Special case for Apache
			if category == "apache" {
				if _, err := os.Stat(checkPath); os.IsNotExist(err) {
					// Fix: checkPath for Apache nested folder
					// checkFile is already "bin/httpd.exe"
					checkPath = filepath.Join(binDir, entry.Name(), "Apache24", checkFile)
				}
			}

			if _, err := os.Stat(checkPath); err == nil {
				// Extract version from folder name
				v := entry.Name()
				v = strings.TrimPrefix(v, "php-")
				v = strings.TrimPrefix(v, "httpd-")
				v = strings.TrimPrefix(v, "mysql-")
				v = strings.TrimPrefix(v, "nginx-")
				v = strings.TrimPrefix(v, "node-v")
				v = strings.TrimPrefix(v, "python-")
				versions[v] = filepath.Join(binDir, entry.Name())
			}
		}
	}
	return versions
}

// GetOpenSSLVersion returns the version of OpenSSL installed in the system.
func GetOpenSSLVersion(exePath string) string {
	cmd := exec.Command(exePath, "version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	parts := strings.Split(string(out), " ")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
