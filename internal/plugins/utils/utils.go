package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetSystemArch returns the architecture string for the current system ("x64" or "x86").
func GetSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

// GetInstalledVersionPaths returns a map of version strings to their absolute paths for a given category.
// It scans the bin/[category] directory for subfolders containing the specified check file.
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

			// Special handling for Apache's directory structure
			if category == "apache" {
				if _, err := os.Stat(checkPath); os.IsNotExist(err) {
					checkPath = filepath.Join(binDir, entry.Name(), "Apache24", checkFile)
				}
			}

			if _, err := os.Stat(checkPath); err == nil {
				// Normalize version string by removing common prefixes
				v := entry.Name()
				v = strings.TrimPrefix(v, "php-")
				v = strings.TrimPrefix(v, "httpd-")
				v = strings.TrimPrefix(v, "mysql-")
				v = strings.TrimPrefix(v, "nginx-")
				v = strings.TrimPrefix(v, "node-v")
				v = strings.TrimPrefix(v, "python-")
				v = strings.TrimPrefix(v, "heidisql-")
				versions[v] = filepath.Join(binDir, entry.Name())
			}
		}
	}
	return versions
}
