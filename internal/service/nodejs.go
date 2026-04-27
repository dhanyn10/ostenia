package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func GetNodeVersion(currentPath string) (string, error) {
	nodeExe := filepath.Join(currentPath, "node.exe")
	if _, err := os.Stat(nodeExe); err != nil {
		return "", err
	}
	cmd := exec.Command(nodeExe, "-v")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
