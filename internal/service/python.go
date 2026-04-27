package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func GetPythonVersion(currentPath string) (string, error) {
	exePath := filepath.Join(currentPath, "python.exe")
	if _, err := os.Stat(exePath); err != nil {
		return "", err
	}
	cmd := exec.Command(exePath, "--version")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
