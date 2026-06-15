package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetSystemDirectory returns the Windows System32 directory or /usr/bin on Unix.
func GetSystemDirectory() string {
	if runtime.GOOS == "windows" {
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		return filepath.Join(systemRoot, "System32")
	}
	return "/usr/bin"
}

// SafeEnv returns a clean environment with a restricted PATH containing only system directories.
func SafeEnv() []string {
	env := os.Environ()
	safePath := "PATH="
	if runtime.GOOS == "windows" {
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		paths := []string{
			filepath.Join(systemRoot, "System32"),
			systemRoot,
			filepath.Join(systemRoot, "System32", "Wbem"),
			filepath.Join(systemRoot, "System32", "WindowsPowerShell", "v1.0"),
		}
		safePath += strings.Join(paths, ";")
	} else {
		safePath += "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	}

	var newEnv []string
	for _, e := range env {
		if !strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			newEnv = append(newEnv, e)
		}
	}
	newEnv = append(newEnv, safePath)
	return newEnv
}
