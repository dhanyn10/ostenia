package utils

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type ModuleDefinition struct {
	Name      string
	CheckFile string
}

// DownloadFile downloads a file from URL to the specified path.
func DownloadFile(path string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// GetSystemArch returns the architecture string for the current system ("x64" or "x86").
func GetSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

// DetectHeidiSQLInstallation checks if HeidiSQL is installed in the system and returns its path and uninstaller.
func DetectHeidiSQLInstallation() (exePath string, uninstaller string) {
	if runtime.GOOS != "windows" {
		return "", ""
	}

	// 1. Check common installation paths
	commonPaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "HeidiSQL", "heidisql.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "HeidiSQL", "heidisql.exe"),
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			exePath = p
			// Inno Setup (HeidiSQL) uses unins000.exe
			uninstaller = filepath.Join(filepath.Dir(p), "unins000.exe")
			if _, err := os.Stat(uninstaller); err != nil {
				uninstaller = filepath.Join(filepath.Dir(p), "uninstall.exe")
				if _, err := os.Stat(uninstaller); err != nil {
					uninstaller = ""
				}
			}
			return
		}
	}

	// 2. Fallback to Registry check via WMIC or cmd
	// We use "where heidisql" as a simple fallback if it's in the PATH
	cmd := exec.Command("cmd", "/c", "where heidisql.exe")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	out, err := cmd.Output()
	if err == nil {
		p := strings.TrimSpace(strings.Split(string(out), "\r\n")[0])
		if p != "" {
			exePath = p
			uninstaller = filepath.Join(filepath.Dir(p), "unins000.exe")
			if _, err := os.Stat(uninstaller); err != nil {
				uninstaller = filepath.Join(filepath.Dir(p), "uninstall.exe")
				if _, err := os.Stat(uninstaller); err != nil {
					uninstaller = ""
				}
			}
			return
		}
	}

	return "", ""
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
