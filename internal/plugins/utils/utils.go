package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func GetSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

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
			// checkFile is usually "bin/httpd.exe" or "php.exe"
			checkPath := filepath.Join(binDir, entry.Name(), checkFile)

			// Special case for Apache where it might be nested inside Apache24 folder
			if category == "apache" {
				if _, err := os.Stat(checkPath); os.IsNotExist(err) {
					// If bin/httpd.exe not found in root, check in Apache24/bin/httpd.exe
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
